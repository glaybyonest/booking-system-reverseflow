export type Session = {
  id: string;
  eventId?: string | null;
  hallId?: string | null;
  startsAt?: string | null;
  endsAt?: string | null;
  status: string;
  isBookable?: boolean;
  externalSource?: string | null;
  externalId?: string | null;
  sourceUrl?: string | null;
  event?: {
    id: string;
    title: string;
  };
  hall?: {
    id?: string | null;
    name: string;
    venue?: string | null;
  };
  hallName?: string | null;
};
