import { create } from "zustand";
import { persist } from "zustand/middleware";

const DEFAULT_VISIBLE_COLUMNS = [
  "num", "track", "artist", "album", "genre", "length", "added", "source",
];

const ALL_COLUMNS = [
  "num", "track", "artist", "album", "genre", "length", "added", "source",
  "popularity", "bpm", "key", "mode", "time_sig", "energy", "danceability",
  "valence", "acousticness", "instrumentalness", "liveness", "speechiness",
  "loudness", "explicit",
];

function defaultVisibility(): Record<string, boolean> {
  return Object.fromEntries(ALL_COLUMNS.map(id => [id, DEFAULT_VISIBLE_COLUMNS.includes(id)]));
}

interface FilterState {
  tracksSearch: string;
  genres: string[];
  yearMin: number | undefined;
  yearMax: number | undefined;
  popularityMin: number | undefined;
  popularityMax: number | undefined;
  durationMin: number | undefined;
  durationMax: number | undefined;
  explicit: boolean | undefined;
  playlistId: string | undefined;
  savedAtMin: string | undefined;
  savedAtMax: string | undefined;
  artistPopularityMin: number | undefined;
  artistPopularityMax: number | undefined;
  artistFollowersMin: number | undefined;
  artistFollowersMax: number | undefined;
  tempoMin: number | undefined;
  tempoMax: number | undefined;
  energyMin: number | undefined;
  energyMax: number | undefined;
  danceabilityMin: number | undefined;
  danceabilityMax: number | undefined;
  valenceMin: number | undefined;
  valenceMax: number | undefined;
  acousticnessMin: number | undefined;
  acousticnessMax: number | undefined;
  instrumentalnessMin: number | undefined;
  instrumentalnessMax: number | undefined;
  livenessMin: number | undefined;
  livenessMax: number | undefined;
  speechinessMin: number | undefined;
  speechinessMax: number | undefined;
  loudnessMin: number | undefined;
  loudnessMax: number | undefined;
  keys: number[];
  mode: number | undefined;
  timeSignatures: number[];
  page: number;
  pageSize: number;
  sortBy: string;
  sortDir: "asc" | "desc";
  tracksColumnVisibility: Record<string, boolean>;
  tracksColumnOrder: string[];

  setTracksSearch: (s: string) => void;
  setGenres: (g: string[]) => void;
  setYearMin: (v: number | undefined) => void;
  setYearMax: (v: number | undefined) => void;
  setPopularityMin: (v: number | undefined) => void;
  setPopularityMax: (v: number | undefined) => void;
  setDurationMin: (v: number | undefined) => void;
  setDurationMax: (v: number | undefined) => void;
  setExplicit: (v: boolean | undefined) => void;
  setPlaylistId: (v: string | undefined) => void;
  setSavedAtMin: (v: string | undefined) => void;
  setSavedAtMax: (v: string | undefined) => void;
  setArtistPopularityMin: (v: number | undefined) => void;
  setArtistPopularityMax: (v: number | undefined) => void;
  setArtistFollowersMin: (v: number | undefined) => void;
  setArtistFollowersMax: (v: number | undefined) => void;
  setTempoMin: (v: number | undefined) => void;
  setTempoMax: (v: number | undefined) => void;
  setEnergyMin: (v: number | undefined) => void;
  setEnergyMax: (v: number | undefined) => void;
  setDanceabilityMin: (v: number | undefined) => void;
  setDanceabilityMax: (v: number | undefined) => void;
  setValenceMin: (v: number | undefined) => void;
  setValenceMax: (v: number | undefined) => void;
  setAcousticnessMin: (v: number | undefined) => void;
  setAcousticnessMax: (v: number | undefined) => void;
  setInstrumentalnessMin: (v: number | undefined) => void;
  setInstrumentalnessMax: (v: number | undefined) => void;
  setLivenessMin: (v: number | undefined) => void;
  setLivenessMax: (v: number | undefined) => void;
  setSpeechinessMin: (v: number | undefined) => void;
  setSpeechinessMax: (v: number | undefined) => void;
  setLoudnessMin: (v: number | undefined) => void;
  setLoudnessMax: (v: number | undefined) => void;
  setKeys: (v: number[]) => void;
  setMode: (v: number | undefined) => void;
  setTimeSignatures: (v: number[]) => void;
  setPage: (p: number) => void;
  setPageSize: (s: number) => void;
  setSortBy: (col: string) => void;
  setSortDir: (dir: "asc" | "desc") => void;
  setTracksColumnVisibility: (id: string, visible: boolean) => void;
  setTracksColumnOrder: (ids: string[]) => void;
  resetTracksColumns: () => void;
  reset: () => void;
}

const filterDefaults = {
  tracksSearch: "",
  genres: [] as string[],
  yearMin: undefined,
  yearMax: undefined,
  popularityMin: undefined,
  popularityMax: undefined,
  durationMin: undefined,
  durationMax: undefined,
  explicit: undefined,
  playlistId: undefined,
  savedAtMin: undefined,
  savedAtMax: undefined,
  artistPopularityMin: undefined,
  artistPopularityMax: undefined,
  artistFollowersMin: undefined,
  artistFollowersMax: undefined,
  tempoMin: undefined,
  tempoMax: undefined,
  energyMin: undefined,
  energyMax: undefined,
  danceabilityMin: undefined,
  danceabilityMax: undefined,
  valenceMin: undefined,
  valenceMax: undefined,
  acousticnessMin: undefined,
  acousticnessMax: undefined,
  instrumentalnessMin: undefined,
  instrumentalnessMax: undefined,
  livenessMin: undefined,
  livenessMax: undefined,
  speechinessMin: undefined,
  speechinessMax: undefined,
  loudnessMin: undefined,
  loudnessMax: undefined,
  keys: [] as number[],
  mode: undefined,
  timeSignatures: [] as number[],
  page: 1,
  pageSize: 50,
  sortBy: "",
  sortDir: "asc" as const,
};

export const useFilterStore = create<FilterState>()(
  persist(
    (set) => ({
      ...filterDefaults,
      tracksColumnVisibility: defaultVisibility(),
      tracksColumnOrder: [...ALL_COLUMNS],

      setTracksSearch: tracksSearch => { set({ tracksSearch, page: 1 }); },
      setGenres: genres => { set({ genres, page: 1 }); },
      setYearMin: yearMin => { set({ yearMin, page: 1 }); },
      setYearMax: yearMax => { set({ yearMax, page: 1 }); },
      setPopularityMin: popularityMin => { set({ popularityMin, page: 1 }); },
      setPopularityMax: popularityMax => { set({ popularityMax, page: 1 }); },
      setDurationMin: durationMin => { set({ durationMin, page: 1 }); },
      setDurationMax: durationMax => { set({ durationMax, page: 1 }); },
      setExplicit: explicit => { set({ explicit, page: 1 }); },
      setPlaylistId: playlistId => { set({ playlistId, page: 1 }); },
      setSavedAtMin: savedAtMin => { set({ savedAtMin, page: 1 }); },
      setSavedAtMax: savedAtMax => { set({ savedAtMax, page: 1 }); },
      setArtistPopularityMin: artistPopularityMin => { set({ artistPopularityMin, page: 1 }); },
      setArtistPopularityMax: artistPopularityMax => { set({ artistPopularityMax, page: 1 }); },
      setArtistFollowersMin: artistFollowersMin => { set({ artistFollowersMin, page: 1 }); },
      setArtistFollowersMax: artistFollowersMax => { set({ artistFollowersMax, page: 1 }); },
      setTempoMin: tempoMin => { set({ tempoMin, page: 1 }); },
      setTempoMax: tempoMax => { set({ tempoMax, page: 1 }); },
      setEnergyMin: energyMin => { set({ energyMin, page: 1 }); },
      setEnergyMax: energyMax => { set({ energyMax, page: 1 }); },
      setDanceabilityMin: danceabilityMin => { set({ danceabilityMin, page: 1 }); },
      setDanceabilityMax: danceabilityMax => { set({ danceabilityMax, page: 1 }); },
      setValenceMin: valenceMin => { set({ valenceMin, page: 1 }); },
      setValenceMax: valenceMax => { set({ valenceMax, page: 1 }); },
      setAcousticnessMin: acousticnessMin => { set({ acousticnessMin, page: 1 }); },
      setAcousticnessMax: acousticnessMax => { set({ acousticnessMax, page: 1 }); },
      setInstrumentalnessMin: instrumentalnessMin => { set({ instrumentalnessMin, page: 1 }); },
      setInstrumentalnessMax: instrumentalnessMax => { set({ instrumentalnessMax, page: 1 }); },
      setLivenessMin: livenessMin => { set({ livenessMin, page: 1 }); },
      setLivenessMax: livenessMax => { set({ livenessMax, page: 1 }); },
      setSpeechinessMin: speechinessMin => { set({ speechinessMin, page: 1 }); },
      setSpeechinessMax: speechinessMax => { set({ speechinessMax, page: 1 }); },
      setLoudnessMin: loudnessMin => { set({ loudnessMin, page: 1 }); },
      setLoudnessMax: loudnessMax => { set({ loudnessMax, page: 1 }); },
      setKeys: keys => { set({ keys, page: 1 }); },
      setMode: mode => { set({ mode, page: 1 }); },
      setTimeSignatures: timeSignatures => { set({ timeSignatures, page: 1 }); },
      setPage: page => { set({ page }); },
      setPageSize: pageSize => { set({ pageSize, page: 1 }); },
      setSortBy: sortBy => { set({ sortBy }); },
      setSortDir: sortDir => { set({ sortDir }); },
      setTracksColumnVisibility: (id, visible) => {
        set(state => ({
          tracksColumnVisibility: { ...state.tracksColumnVisibility, [id]: visible },
        }));
      },
      setTracksColumnOrder: ids => { set({ tracksColumnOrder: ids }); },
      resetTracksColumns: () => {
        set({
          tracksColumnVisibility: defaultVisibility(),
          tracksColumnOrder: [...ALL_COLUMNS],
        });
      },
      reset: () => { set({ ...filterDefaults, page: 1 }); },
    }),
    {
      name: "spotifind.filters",
      partialize: (state) => ({
        tracksSearch: state.tracksSearch,
        genres: state.genres,
        yearMin: state.yearMin,
        yearMax: state.yearMax,
        popularityMin: state.popularityMin,
        popularityMax: state.popularityMax,
        durationMin: state.durationMin,
        durationMax: state.durationMax,
        explicit: state.explicit,
        playlistId: state.playlistId,
        savedAtMin: state.savedAtMin,
        savedAtMax: state.savedAtMax,
        artistPopularityMin: state.artistPopularityMin,
        artistPopularityMax: state.artistPopularityMax,
        artistFollowersMin: state.artistFollowersMin,
        artistFollowersMax: state.artistFollowersMax,
        tempoMin: state.tempoMin,
        tempoMax: state.tempoMax,
        energyMin: state.energyMin,
        energyMax: state.energyMax,
        danceabilityMin: state.danceabilityMin,
        danceabilityMax: state.danceabilityMax,
        valenceMin: state.valenceMin,
        valenceMax: state.valenceMax,
        acousticnessMin: state.acousticnessMin,
        acousticnessMax: state.acousticnessMax,
        instrumentalnessMin: state.instrumentalnessMin,
        instrumentalnessMax: state.instrumentalnessMax,
        livenessMin: state.livenessMin,
        livenessMax: state.livenessMax,
        speechinessMin: state.speechinessMin,
        speechinessMax: state.speechinessMax,
        loudnessMin: state.loudnessMin,
        loudnessMax: state.loudnessMax,
        keys: state.keys,
        mode: state.mode,
        timeSignatures: state.timeSignatures,
        pageSize: state.pageSize,
        sortBy: state.sortBy,
        sortDir: state.sortDir,
        tracksColumnVisibility: state.tracksColumnVisibility,
        tracksColumnOrder: state.tracksColumnOrder,
      }),
    }
  )
);
