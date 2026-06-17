import { useEffect, useState } from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, BookOpen, FileText, Globe, Clock, User } from "lucide-react";

import { GetBook } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/service";
import type { BookDetail } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";

export const Route = createFileRoute("/books/$dirName")({
	component: BookPage,
});

function BookPage() {
	const { dirName } = Route.useParams();
	const [book, setBook] = useState<BookDetail | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		GetBook(dirName)
			.then(setBook)
			.catch((err: unknown) => {
				const msg = err instanceof Error ? err.message : String(err);
				setError(msg);
			})
			.finally(() => setLoading(false));
	}, [dirName]);

	if (loading) {
		return (
			<div className="mx-auto flex max-w-2xl p-6">
				<p className="text-sm text-muted-foreground">Loading...</p>
			</div>
		);
	}

	if (error || !book) {
		return (
			<div className="mx-auto flex max-w-2xl flex-col gap-4 p-6">
				<div className="border border-destructive/50 bg-destructive/10 px-4 py-2 text-xs text-destructive">
					{error || "Book not found"}
				</div>
				<Link
					to="/"
					className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
				>
					<ArrowLeft className="size-4" />
					Back to library
				</Link>
			</div>
		);
	}

	return (
		<div className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
			<Link
				to="/"
				className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
			>
				<ArrowLeft className="size-4" />
				Back to library
			</Link>

			<div className="flex items-center gap-3">
				<BookOpen className="size-6" />
				<h1 className="text-2xl font-bold">{book.metadata.title}</h1>
			</div>

			<div className="flex flex-col gap-2 border border-border bg-muted/30 px-4 py-3 text-xs">
				{book.metadata.author && (
					<div className="flex items-center gap-2">
						<User className="size-3.5 text-muted-foreground" />
						<span className="text-muted-foreground">Author:</span>
						<span>{book.metadata.author}</span>
					</div>
				)}
				<div className="flex items-center gap-2">
					<Globe className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Language:</span>
					<span>{book.metadata.language}</span>
				</div>
				<div className="flex items-center gap-2">
					<FileText className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Source:</span>
					<span>{book.metadata.sourceFile}</span>
				</div>
				<div className="flex items-center gap-2">
					<Clock className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Imported:</span>
					<span>{new Date(book.metadata.importedAt).toLocaleString()}</span>
				</div>
				<div className="mt-1 flex items-center gap-2 border-t border-border pt-2">
					<span className="text-muted-foreground">Status:</span>
					<span>{book.state.status}</span>
				</div>
			</div>
		</div>
	);
}
