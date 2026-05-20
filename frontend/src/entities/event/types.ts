import type { Session } from "@/entities/session/types";

export type EventVenue = {
  id: string;
  name: string;
  address: string;
  city: string;
  latitude?: number | null;
  longitude?: number | null;
  metroStations?: string[];
  venueTypeCode?: string | null;
  venueTypeName?: string | null;
};

export type EventExternalLink = {
  id: string;
  externalSource: string;
  externalId: string;
  sourceUrl?: string | null;
  importedAt?: string | null;
};

export type Event = {
  id: string;
  title: string;
  description?: string | null;
  longDescription?: string | null;
  category?: string | null;
  posterUrl?: string | null;
  status: string;
  source: string;
  externalSource?: string | null;
  sourceUrl?: string | null;
  bookingMode: string;
  startsAt?: string | null;
  endsAt?: string | null;
  ageRestriction?: string | null;
  priceMin?: number | null;
  priceMax?: number | null;
  tags?: string[];
  ratingCount?: number | null;
  isImported?: boolean;
  venue?: EventVenue | null;
  externalLinks?: EventExternalLink[];
  sessions?: Session[];
  createdAt?: string | null;
  updatedAt?: string | null;
};
