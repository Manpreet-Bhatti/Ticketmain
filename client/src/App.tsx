import { useEffect, useState } from "react";
import Venue from "./components/Venue";
import layoutData from "../../venue_layout.json";
import type { VenueLayout } from "./types";
import { oneWeekFromTodayAt8PM } from "./utils/dateOneWeek";

const venueLayout = layoutData as VenueLayout;

function App() {
  const [userId, setUserId] = useState<string>("");

  useEffect(() => {
    let storedId = localStorage.getItem("userId");
    if (!storedId) {
      storedId = `user_${Math.random().toString(36).substring(2, 11)}`;
      localStorage.setItem("userId", storedId);
    }
    setUserId(storedId);
  }, []);

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
