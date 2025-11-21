import { useEffect, useState } from "react";
import Venue from "./components/Venue";
import type { VenueLayout } from "./types";
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
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:3000";
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
  }, []);

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
        <Venue layout={venueLayout} />
      </main>
    </div>
  );
}

export default App;
