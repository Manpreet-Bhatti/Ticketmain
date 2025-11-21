import React from "react";
import type { VenueLayout, Section, SeatState } from "../types";

interface VenueProps {
  layout: VenueLayout;
  seatStates: SeatState[];
  onSeatClick: (seatId: string) => void;
}

const StarSolid = ({ className, ...props }: React.SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="currentColor"
    className={className}
    {...props}
  >
    <path
      fillRule="evenodd"
      d="M10.788 3.21c.448-1.077 1.976-1.077 2.424 0l2.082 5.007 5.404.433c1.164.093 1.636 1.545.749 2.305l-4.117 3.527 1.257 5.273c.271 1.136-.964 2.033-1.96 1.425L12 18.354 7.373 21.18c-.996.608-2.231-.29-1.96-1.425l1.257-5.273-4.117-3.527c-.887-.76-.415-2.212.749-2.305l5.404-.433 2.082-5.006z"
      clipRule="evenodd"
    />
  </svg>
);

const CheckIcon = ({ className, ...props }: React.SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="currentColor"
    className={className}
    {...props}
  >
    <path
      fillRule="evenodd"
      d="M19.916 4.626a.75.75 0 01.208 1.04l-9 13.5a.75.75 0 01-1.154.114l-6-6a.75.75 0 011.06-1.06l5.353 5.353 8.493-12.739a.75.75 0 011.04-.208z"
      clipRule="evenodd"
    />
  </svg>
);

const Venue: React.FC<VenueProps> = ({ layout, seatStates, onSeatClick }) => {
  const { dimensions, stage_area, sections } = layout;
  const { rows, cols } = dimensions;

  const getSection = (r: number, c: number): Section | undefined => {
    return sections.find(
      (s) =>
        r >= s.row_start && r <= s.row_end && c >= s.col_start && c <= s.col_end
    );
  };

  const isStage = (r: number, c: number): boolean => {
    return (
      r >= stage_area.row_start &&
      r <= stage_area.row_end &&
      c >= stage_area.col_start &&
      c <= stage_area.col_end
    );
  };

  const renderCell = (r: number, c: number) => {
    // Check for Stage
    if (isStage(r, c)) {
      return <div key={`${r}-${c}`} className="w-full h-full bg-black" />;
    }

    // Check for Section (Seat)
    const section = getSection(r, c);
    if (section) {
      const seatId = `${r}-${c}`;
      const seatState = seatStates.find((s) => s.seatId === seatId);
      const isHeld = seatState?.status === "HELD";

      let sectionColorClass = "";
      if (isHeld) {
        sectionColorClass = "bg-green-500 hover:bg-green-600";
      } else {
        switch (section.id) {
          case "vip_left_wing":
          case "vip_right_wing":
          case "vip_center":
            sectionColorClass = "bg-yellow-500 hover:bg-yellow-600 ";
            break;
          default:
            sectionColorClass = "bg-blue-500 hover:bg-blue-600";
        }
      }

      return (
        <div
          key={`${r}-${c}`}
          className={`w-full h-full rounded cursor-pointer transition-transform border-2 border-white flex items-center justify-center ${sectionColorClass}`}
          onClick={() => onSeatClick(seatId)}
          title={`${section.name} - $${section.price}`}
        >
          {isHeld ? (
            <CheckIcon className="w-4 h-4 text-white" />
          ) : (
            (section.id === "vip_left_wing" ||
              section.id === "vip_right_wing" ||
              section.id === "vip_center") && <StarSolid className="w-4 h-4" />
          )}
        </div>
      );
    }

    // Empty space
    return <div key={`${r}-${c}`} className="w-full h-full bg-transparent" />;
  };

  const gridItems = [];
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      gridItems.push(renderCell(r, c));
    }
  }

  return (
    <div className="inline-block bg-white p-3 rounded-lg relative">
      <div
        className="relative"
        style={{
          display: "grid",
          gridTemplateColumns: `repeat(${cols}, 20px)`,
          gridTemplateRows: `repeat(${rows}, 20px)`,
          justifyContent: "center",
        }}
      >
        {gridItems}

        {/* Overlay for Stage Label */}
        <div
          className="absolute flex items-center justify-center text-white font-bold text-2xl pointer-events-none z-10"
          style={{
            gridColumnStart: stage_area.col_start + 8,
            gridColumnEnd: stage_area.col_end + 9,
            gridRowStart: stage_area.row_start + 5,
            gridRowEnd: stage_area.row_end + 6,
          }}
        >
          {stage_area.label}
        </div>
      </div>
    </div>
  );
};

export default Venue;
