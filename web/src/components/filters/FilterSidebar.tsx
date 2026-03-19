import { useQuery } from "@tanstack/react-query";
import { X } from "lucide-react";
import { getGenres } from "../../api/sync";
import { useFilterStore } from "../../stores/filterStore";

export default function FilterSidebar() {
  const { data: allGenres = [] } = useQuery({ queryKey: ["genres"], queryFn: getGenres });
  const {
    genres,
    setGenres,
    yearMin,
    yearMax,
    setYearMin,
    setYearMax,
    popularityMin,
    popularityMax,
    setPopularityMin,
    setPopularityMax,
    explicit,
    setExplicit,
    reset,
  } = useFilterStore();

  function toggleGenre(g: string) {
    if (genres.includes(g)) {
      setGenres(genres.filter(x => x !== g));
    } else {
      setGenres([...genres, g]);
    }
  }

  const hasFilters =
    genres.length > 0
    || yearMin != null
    || yearMax != null
    || popularityMin != null
    || popularityMax != null
    || explicit != null;

  return (
    <div className="w-56 shrink-0 border-r border-gray-800 overflow-y-auto p-4 space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-300">Filters</h3>
        {hasFilters && (
          <button
            onClick={reset}
            className="flex items-center gap-1 text-xs text-gray-500 hover:text-white"
          >
            <X className="w-3 h-3" /> Clear
          </button>
        )}
      </div>

      <section className="space-y-2">
        <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide">Year</h4>
        <div className="flex gap-2">
          <input
            type="number"
            placeholder="From"
            value={yearMin ?? ""}
            onChange={e => setYearMin(e.target.value ? Number(e.target.value) : undefined)}
            className="w-full px-2 py-1 bg-gray-900 border border-gray-700 rounded text-sm text-white placeholder-gray-600 focus:outline-none focus:border-gray-500"
          />
          <input
            type="number"
            placeholder="To"
            value={yearMax ?? ""}
            onChange={e => setYearMax(e.target.value ? Number(e.target.value) : undefined)}
            className="w-full px-2 py-1 bg-gray-900 border border-gray-700 rounded text-sm text-white placeholder-gray-600 focus:outline-none focus:border-gray-500"
          />
        </div>
      </section>

      <section className="space-y-2">
        <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide">Popularity</h4>
        <div className="flex gap-2">
          <input
            type="number"
            min={0}
            max={100}
            placeholder="Min"
            value={popularityMin ?? ""}
            onChange={e => setPopularityMin(e.target.value ? Number(e.target.value) : undefined)}
            className="w-full px-2 py-1 bg-gray-900 border border-gray-700 rounded text-sm text-white placeholder-gray-600 focus:outline-none focus:border-gray-500"
          />
          <input
            type="number"
            min={0}
            max={100}
            placeholder="Max"
            value={popularityMax ?? ""}
            onChange={e => setPopularityMax(e.target.value ? Number(e.target.value) : undefined)}
            className="w-full px-2 py-1 bg-gray-900 border border-gray-700 rounded text-sm text-white placeholder-gray-600 focus:outline-none focus:border-gray-500"
          />
        </div>
      </section>

      <section className="space-y-2">
        <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide">Explicit</h4>
        <div className="flex gap-2">
          {[undefined, true, false].map(val => (
            <button
              key={String(val)}
              onClick={() => setExplicit(val)}
              className={`flex-1 py-1 text-xs rounded border transition-colors ${
                explicit === val
                  ? "border-[#1DB954] text-[#1DB954] bg-[#1DB954]/10"
                  : "border-gray-700 text-gray-400 hover:border-gray-500"
              }`}
            >
              {val === undefined ? "All" : val ? "Yes" : "No"}
            </button>
          ))}
        </div>
      </section>

      {allGenres.length > 0 && (
        <section className="space-y-2">
          <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide">Genre</h4>
          <div className="flex flex-wrap gap-1.5 max-h-64 overflow-y-auto">
            {allGenres.map(g => (
              <button
                key={g}
                onClick={() => toggleGenre(g)}
                className={`px-2 py-0.5 text-xs rounded-full border transition-colors ${
                  genres.includes(g)
                    ? "border-[#1DB954] text-[#1DB954] bg-[#1DB954]/10"
                    : "border-gray-700 text-gray-400 hover:border-gray-500"
                }`}
              >
                {g}
              </button>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
