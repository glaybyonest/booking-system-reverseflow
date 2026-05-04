export type Session = {
  id: string;
  eventId?: string;
  hallId?: string;
  startsAt?: string;
  endsAt?: string;
  status: string;
  event?: {
    id: string;
    title: string;
  };
  hall?: {
    id?: string;
    name: string;
    venue?: string;
  };
  hallName?: string;
};
