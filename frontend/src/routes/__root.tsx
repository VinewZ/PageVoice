import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { Home } from "lucide-react";

import { ModeToggle } from "@/components/mode-toggle";
import { ThemeProvider } from "@/components/theme-provider";
import { buttonVariants } from "@/components/ui/button";

import "../styles.css";
import { Devtools } from "@/lib/devtools";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <header className="flex items-center justify-between border-b border-border px-4 py-2">
        <nav className="flex items-center gap-2">
          <Link
            to="/"
            className={buttonVariants({ variant: "ghost", size: "sm" })}
          >
            <Home className="size-4" />
            Home
          </Link>
        </nav>
        <ModeToggle />
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
      <Devtools />
    </ThemeProvider>
  );
}
