const fallbackLat = 55.751244;
const fallbackLon = 37.618423;
const fallbackZoom = 11;

export function defaultMapCenter() {
  return {
    lat: parseNumber(process.env.NEXT_PUBLIC_MAP_DEFAULT_LAT, fallbackLat),
    lon: parseNumber(process.env.NEXT_PUBLIC_MAP_DEFAULT_LON, fallbackLon),
    zoom: parseNumber(process.env.NEXT_PUBLIC_MAP_DEFAULT_ZOOM, fallbackZoom)
  };
}

function parseNumber(value: string | undefined, fallback: number) {
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}
