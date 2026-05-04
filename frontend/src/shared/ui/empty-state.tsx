import { Card, CardContent } from "@/shared/ui/card";

export function EmptyState({ title, description }: { title: string; description?: string }) {
  return (
    <Card>
      <CardContent className="text-center">
        <h3 className="text-lg font-bold">{title}</h3>
        {description ? <p className="mt-2 text-sm text-gray-500">{description}</p> : null}
      </CardContent>
    </Card>
  );
}
