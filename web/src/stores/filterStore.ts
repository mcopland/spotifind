import { create } from "zustand";

interface FilterState {
  tracksSearch: string;
  genres: string[];
  yearMin: number | undefined;
  yearMax: number | undefined;
  popularityMin: number | undefined;
  popularityMax: number | undefined;
  explicit: boolean | undefined;
  page: number;
  pageSize: number;
  sortBy: string;
  sortDir: "asc" | "desc";

  setTracksSearch: (s: string) => void;
  setGenres: (g: string[]) => void;
  setYearMin: (v: number | undefined) => void;
  setYearMax: (v: number | undefined) => void;
  setPopularityMin: (v: number | undefined) => void;
  setPopularityMax: (v: number | undefined) => void;
  setExplicit: (v: boolean | undefined) => void;
  setPage: (p: number) => void;
  setPageSize: (s: number) => void;
  setSortBy: (col: string) => void;
  setSortDir: (dir: "asc" | "desc") => void;
  reset: () => void;
}

const defaults = {
  tracksSearch: "",
  genres: [] as string[],
  yearMin: undefined,
  yearMax: undefined,
  popularityMin: undefined,
  popularityMax: undefined,
  explicit: undefined,
  page: 1,
  pageSize: 50,
  sortBy: "",
  sortDir: "asc" as const,
};

export const useFilterStore = create<FilterState>(set => ({
  ...defaults,
  setTracksSearch: tracksSearch => { set({ tracksSearch, page: 1 }); },
  setGenres: genres => { set({ genres, page: 1 }); },
  setYearMin: yearMin => { set({ yearMin, page: 1 }); },
  setYearMax: yearMax => { set({ yearMax, page: 1 }); },
  setPopularityMin: popularityMin => { set({ popularityMin, page: 1 }); },
  setPopularityMax: popularityMax => { set({ popularityMax, page: 1 }); },
  setExplicit: explicit => { set({ explicit, page: 1 }); },
  setPage: page => { set({ page }); },
  setPageSize: pageSize => { set({ pageSize, page: 1 }); },
  setSortBy: sortBy => { set({ sortBy }); },
  setSortDir: sortDir => { set({ sortDir }); },
  reset: () => { set(defaults); },
}));
