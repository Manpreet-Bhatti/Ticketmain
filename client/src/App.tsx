import { useEffect, useState } from "react";
import Venue from "./components/Venue";
import type { VenueLayout, SeatState } from "./types";
import { oneWeekFromTodayAt8PM } from "./utils/dateOneWeek";

function App() {
  const [userId] = useState<string>(() => {
    const stored = localStorage.getItem("userId");
    if (stored) return stored;

    const newId = `user_${Math.random().toString(36).substring(2, 11)}`;
    localStorage.setItem("userId", newId);
    return newId;
  });
  const [venueLayout, setVenueLayout] = useState<VenueLayout | null>(null);
  const [seatStates, setSeatStates] = useState<SeatState[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:3000";

  const fetchSeats = () => {
    fetch(`${apiUrl}/api/seats`)
      .then((res) => res.json())
      .then((data) => setSeatStates(data))
      .catch((err) => console.error("Error fetching seats:", err));
  };

  useEffect(() => {
    fetch(`${apiUrl}/api/venue`)
      .then((res) => {
        if (!res.ok) {
          throw new Error("Failed to fetch seating map");
        }
        return res.json();
      })
      .then((data) => {
        setVenueLayout(data as VenueLayout);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Error fetching seating map:", err);
        setError(err.message);
        setLoading(false);
      });

    fetchSeats();

    const wsUrl = apiUrl.replace("http", "ws") + "/ws";
    const socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log("Connected to WebSocket");
    };

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "SEAT_UPDATE") {
        const { seatId, status, ownerId } = message.payload;
        setSeatStates((prev) => {
          const filtered = prev.filter((s) => s.seatId !== seatId);

          return [...filtered, { seatId, status, ownerId }];
        });
      }
    };

    socket.onclose = () => {
      console.log("Disconnected from WebSocket");
    };

    return () => {
      socket.close();
    };
  }, [apiUrl]);

  const handleSeatClick = (seatId: string) => {
    const seatState = seatStates.find((s) => s.seatId === seatId);
    const isHeld = seatState?.status === "HELD";

    if (isHeld) {
      if (seatState?.ownerId === userId) {
        fetch(`${apiUrl}/api/hold`, {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ seatId, userId }),
        })
          .then((res) => res.json())
          .then((data) => {
            if (data.status !== "success") {
              alert(data.message);
            }
          })
          .catch((err) => console.error("Error releasing seat:", err));
      } else {
        alert("Seat already taken by another user");
      }
      return;
    }

    fetch(`${apiUrl}/api/hold`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ seatId, userId }),
    })
      .then((res) => {
        if (res.status === 409) {
          throw new Error("Seat already taken");
        }
        return res.json();
      })
      .then(() => {
        // WebSocket will handle update
      })
      .catch((err) => {
        alert(err.message || "Error holding seat");
        console.error("Error holding seat:", err);
      });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">
        Loading seating map...
      </div>
    );
  }

  if (error || !venueLayout) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900 text-red-500">
        Error: {error || "Seating map not found"}
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center min-h-screen bg-gray-200 text-black p-8">
      <header className="mb-6 w-full text-center">
        <p className="text-black text-right w-full">User: {userId}</p>
        <h1 className="text-4xl font-bold mb-2">{venueLayout.venue_name}</h1>
        <h2 className="text-2xl font-semibold mb-2">
          {oneWeekFromTodayAt8PM()}
        </h2>
        <h2 className="text-2xl font-semibold underline">
          {venueLayout.venue_location}
        </h2>
      </header>
      <main>
        <Venue
          layout={venueLayout}
          seatStates={seatStates}
          onSeatClick={handleSeatClick}
        />
      </main>
    </div>
  );
}

export default App;
