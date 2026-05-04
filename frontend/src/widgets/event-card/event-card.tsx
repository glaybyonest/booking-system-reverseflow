import Link from "next/link";

import type { Event } from "@/entities/event/types";
import { routes } from "@/shared/lib/routes";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";

export function EventCard({ event }: { event: Event }) {
  return (
    <Card className="transition-all hover:-translate-y-0.5 hover:shadow-md">
      <CardContent className="flex h-full flex-col">
        <div className="mb-5 flex items-center justify-between gap-3">
          <Badge>{event.category ?? "Мероприятие"}</Badge>
          <span className="text-xs font-medium text-gray-400">{event.status}</span>
        </div>
        <h3 className="text-2xl font-bold tracking-tight">{event.title}</h3>
        <p className="mt-3 line-clamp-3 text-sm leading-relaxed text-gray-500">
          {event.description ?? "Описание появится позже."}
        </p>
        <Link href={routes.event(event.id)} className="mt-8">
          <Button className="w-full rounded-full">Открыть</Button>
        </Link>
      </CardContent>
    </Card>
  );
}
