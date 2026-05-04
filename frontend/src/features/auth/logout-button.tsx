"use client";

import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

import { Button } from "@/shared/ui/button";
import { useLogout } from "@/features/auth/auth.hooks";

export function LogoutButton() {
  const router = useRouter();
  const logout = useLogout();
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() =>
        logout.mutate(undefined, {
          onSuccess: () => router.push("/login")
        })
      }
      disabled={logout.isPending}
      aria-label="Выйти"
    >
      <LogOut className="h-4 w-4" />
      <span className="hidden sm:inline">Выйти</span>
    </Button>
  );
}
