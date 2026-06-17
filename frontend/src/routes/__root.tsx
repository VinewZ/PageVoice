import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { FileText, Home } from "lucide-react";

import { ModeToggle } from "#/components/mode-toggle";
import { ThemeProvider } from "#/components/theme-provider";
import { buttonVariants } from "#/components/ui/button";

import "../styles.css";

export const Route = createRootRoute({
	component: RootComponent,
});

function RootComponent() {
	return (
		<ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
			<div className="flex min-h-screen flex-col">
				<header className="flex items-center justify-between border-b border-border px-4 py-2">
					<nav className="flex items-center gap-2">
						<Link
							to="/"
							className={buttonVariants({ variant: "ghost", size: "sm" })}
						>
							<Home className="size-4" />
							Home
						</Link>
						<Link
							to="/upload"
							className={buttonVariants({ variant: "ghost", size: "sm" })}
						>
							<FileText className="size-4" />
							Upload
						</Link>
					</nav>
					<ModeToggle />
				</header>
				<main className="flex-1">
					<Outlet />
				</main>
			</div>
			<TanStackRouterDevtoolsPanel />
		</ThemeProvider>
	);
}
