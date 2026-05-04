import { ArrowRight } from "lucide-react";
import Link from "next/link";

import type { Session } from "@/entities/session/types";
import { formatDateTime, formatTimeRange } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";

export function SessionCard({ session }: { session: Session }) {
  return (
    <Card className="transition-all hover:shadow-md">
      <CardContent className="flex flex-col gap-5 md:flex-row md:items-center md:justify-between">
        <div>
          <Badge variant={session.status === "scheduled" ? "default" : "muted"}>
            {session.status === "scheduled" ? "Запланировано" : session.status}
          </Badge>
          <h3 className="mt-3 text-xl font-bold">{formatDateTime(session.startsAt)}</h3>
          <p className="mt-1 text-sm text-gray-500">
            {formatTimeRange(session.startsAt, session.endsAt)} ·{" "}
            {session.hall?.name ?? session.hallName ?? "Зал уточняется"}
          </p>
        </div>
        <Link href={routes.session(session.id)}>
          <Button className="w-full rounded-full md:w-auto">
            Выбрать места <ArrowRight className="h-4 w-4" />
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
}
