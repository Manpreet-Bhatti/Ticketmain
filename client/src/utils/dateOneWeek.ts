const dateFormatter = new Intl.DateTimeFormat("en-US", {
  weekday: "short",
  month: "short",
  day: "2-digit",
  year: "numeric",
});

const timeFormatter = new Intl.DateTimeFormat("en-US", {
  hour: "numeric",
  minute: "2-digit",
  hour12: true,
});

export function oneWeekFromTodayAt8PM(): string {
  const nextWeek = new Date();
  nextWeek.setDate(nextWeek.getDate() + 7);
  nextWeek.setHours(20, 0, 0, 0);

  const datePart = dateFormatter.format(nextWeek).replace(",", " •");
  return `${datePart} • ${timeFormatter.format(nextWeek)}`;
}
