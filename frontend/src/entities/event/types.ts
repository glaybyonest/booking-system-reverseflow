export type Event = {
  id: string;
  title: string;
  description?: string | null;
  category?: string | null;
  posterUrl?: string | null;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};
