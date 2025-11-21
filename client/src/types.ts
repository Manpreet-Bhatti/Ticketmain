export interface Section {
  id: string;
  name: string;
  price: number;
  row_start: number;
  row_end: number;
  col_start: number;
  col_end: number;
}

export interface VenueLayout {
  venue_location: string;
  venue_name: string;
  dimensions: { rows: number; cols: number };
  stage_area: {
    row_start: number;
    row_end: number;
    col_start: number;
    col_end: number;
    label: string;
  };
  sections: Section[];
}

export type SeatStatus = "AVAILABLE" | "HELD" | "SOLD";

export interface SeatState {
  id: string;
  status: SeatStatus;
  ownerId?: string;
}
