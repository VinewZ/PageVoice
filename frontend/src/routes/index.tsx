import { useEffect, useState } from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import { BookOpen, Library } from "lucide-react";

import { GetLibrary } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/service";
import type { LibraryEntry } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
	const [entries, setEntries] = useState<LibraryEntry[]>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		GetLibrary()
			.then(setEntries)
			.catch(() => setEntries([]))
			.finally(() => setLoading(false));
	}, []);

	return (
		<div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
			<div className="flex items-center gap-3">
				<Library className="size-6" />
				<h1 className="text-2xl font-bold">My Library</h1>
			</div>

			{loading && (
				<p className="text-sm text-muted-foreground">Loading...</p>
			)}

			{!loading && entries.length === 0 && (
				<div className="flex flex-col items-center gap-4 py-16">
					<BookOpen className="size-12 text-muted-foreground/50" />
					<p className="text-sm text-muted-foreground">
						No books imported yet.
					</p>
					<Link
						to="/upload"
						className="text-sm font-medium text-primary underline-offset-4 hover:underline"
					>
						Upload your first document
					</Link>
				</div>
			)}

			{!loading && entries.length > 0 && (
				<div className="flex flex-col gap-1.5">
					{entries.map((entry) => (
						<Link
							key={entry.id}
							to="/books/$dirName"
							params={{ dirName: entry.dirName }}
							className="flex items-center gap-3 border-l-2 border-muted-foreground/20 px-4 py-3 hover:border-primary/40"
						>
							<BookOpen className="size-4 shrink-0 text-muted-foreground" />
							<span className="text-sm">{entry.title}</span>
							<span className="ml-auto text-xs text-muted-foreground">
								{entry.id}
							</span>
						</Link>
					))}
				</div>
			)}
		</div>
	);
}
