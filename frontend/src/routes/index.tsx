import type { UploadResult } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";
import { GetLibrary, ProcessFile } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/service";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
	Book,
	BookOpen,
	CheckCircle2,
	File,
	FileText,
	Library,
	Loader2,
	Upload,
} from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import type { LibraryEntry } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";

export const Route = createFileRoute("/")({ component: Home });

const LANGUAGES = [
	{ value: "czech", label: "Czech" },
	{ value: "danish", label: "Danish" },
	{ value: "dutch", label: "Dutch" },
	{ value: "english", label: "English" },
	{ value: "estonian", label: "Estonian" },
	{ value: "finnish", label: "Finnish" },
	{ value: "french", label: "French" },
	{ value: "german", label: "German" },
	{ value: "greek", label: "Greek" },
	{ value: "italian", label: "Italian" },
	{ value: "norwegian", label: "Norwegian" },
	{ value: "polish", label: "Polish" },
	{ value: "portuguese", label: "Portuguese" },
	{ value: "slovene", label: "Slovene" },
	{ value: "spanish", label: "Spanish" },
	{ value: "swedish", label: "Swedish" },
	{ value: "turkish", label: "Turkish" },
];

const ACCEPTED_TYPES = [".pdf", ".epub", ".txt"];

function fileIcon(type: string) {
	switch (type) {
		case ".pdf":
			return <FileText className="size-8 text-red-500" />;
		case ".epub":
			return <Book className="size-8 text-orange-500" />;
		case ".txt":
			return <File className="size-8 text-blue-500" />;
	}
	return <File className="size-8" />;
}

function Home() {
	const [entries, setEntries] = useState<LibraryEntry[]>([]);
	const [loading, setLoading] = useState(true);

	const [file, setFile] = useState<File | null>(null);
	const [language, setLanguage] = useState("english");
	const [processing, setProcessing] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [result, setResult] = useState<UploadResult | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const [dragging, setDragging] = useState(false);

	const refreshLibrary = useCallback(() => {
		GetLibrary()
			.then(setEntries)
			.catch(() => setEntries([]))
			.finally(() => setLoading(false));
	}, []);

	useEffect(() => {
		refreshLibrary();
	}, [refreshLibrary]);

	const handleFile = useCallback((f: File) => {
		setError(null);
		setResult(null);
		const ext = `.${f.name.split(".").pop()?.toLowerCase()}`;
		if (!ACCEPTED_TYPES.includes(ext)) {
			setError("Only PDF, EPUB, and TXT files are supported.");
			setFile(null);
			return;
		}
		setFile(f);
	}, []);

	const handleDrop = useCallback(
		(e: React.DragEvent) => {
			e.preventDefault();
			setDragging(false);
			const f = e.dataTransfer.files[0];
			if (f) handleFile(f);
		},
		[handleFile],
	);

	const handleDragOver = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		setDragging(true);
	}, []);

	const handleDragLeave = useCallback(() => setDragging(false), []);

	const handleProcess = async () => {
		if (!file) return;
		setProcessing(true);
		setError(null);
		setResult(null);
		try {
			const buffer = await file.arrayBuffer();
			const bytes = new Uint8Array(buffer);
			let binary = "";
			for (let i = 0; i < bytes.length; i++) {
				binary += String.fromCharCode(bytes[i]);
			}
			const base64 = btoa(binary);
			const res = await ProcessFile(file.name, base64, language);
			setResult(res);
			refreshLibrary();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		} finally {
			setProcessing(false);
		}
	};

	return (
		<div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
			{/* Upload section */}
			<div className="flex flex-col gap-4">
				<h1 className="text-2xl font-bold">Upload Document</h1>

				<label
					onDrop={handleDrop}
					onDragOver={handleDragOver}
					onDragLeave={handleDragLeave}
					className={`flex cursor-pointer flex-col items-center justify-center gap-3 rounded-none border-2 border-dashed p-12 transition-colors ${
						dragging
							? "border-primary bg-primary/5"
							: "border-muted-foreground/30 hover:border-muted-foreground/50"
					}`}
				>
					<input
						ref={inputRef}
						type="file"
						accept=".pdf,.epub,.txt"
						className="hidden"
						onChange={(e) => {
							const f = e.target.files?.[0];
							if (f) handleFile(f);
						}}
					/>
					{file ? (
						<>
							{fileIcon(`.${file.name.split(".").pop()?.toLowerCase()}`)}
							<p className="text-sm font-medium">{file.name}</p>
							<p className="text-xs text-muted-foreground">
								{(file.size / 1024).toFixed(1)} KB
							</p>
							<Button
								variant="outline"
								size="xs"
								onClick={(e) => {
									e.stopPropagation();
									setFile(null);
									setResult(null);
								}}
							>
								Remove
							</Button>
						</>
					) : (
						<>
							<Upload className="size-10 text-muted-foreground" />
							<p className="text-sm text-muted-foreground">
								Drop your file here or click to browse
							</p>
							<p className="text-xs text-muted-foreground">
								Supports PDF, EPUB, TXT
							</p>
						</>
					)}
				</label>

				{/* Language + Process */}
				<div className="flex items-end gap-4">
					<div className="flex flex-col gap-1.5">
						<span className="text-xs font-medium text-muted-foreground">
							Language
						</span>
						<Select value={language} onValueChange={(v) => setLanguage(v ?? "")}>
							<SelectTrigger className="w-[140px]">
								<SelectValue placeholder="Select language" />
							</SelectTrigger>
							<SelectContent>
								<SelectGroup>
									{LANGUAGES.map((l) => (
										<SelectItem key={l.value} value={l.value}>
											{l.label}
										</SelectItem>
									))}
								</SelectGroup>
							</SelectContent>
						</Select>
					</div>
					<Button onClick={handleProcess} disabled={!file || processing}>
						{processing ? (
							<>
								<Loader2 className="size-4 animate-spin" />
								Processing...
							</>
						) : (
							"Process File"
						)}
					</Button>
				</div>

				{/* Error */}
				{error && (
					<div className="border border-destructive/50 bg-destructive/10 px-4 py-2 text-xs text-destructive">
						{error}
					</div>
				)}

				{/* Result */}
				{result && (
					<div className="flex flex-col gap-4">
						<div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
							<CheckCircle2 className="size-4" />
							Imported {result.totalChars.toLocaleString()} characters
						</div>

						{(result.metadata.title || result.metadata.author) && (
							<div className="border border-border bg-muted/30 px-4 py-3 text-xs">
								{result.metadata.title && (
									<p>
										<span className="font-medium text-muted-foreground">
											Title:
										</span>{" "}
										{result.metadata.title}
									</p>
								)}
								{result.metadata.author && (
									<p>
										<span className="font-medium text-muted-foreground">
											Author:
										</span>{" "}
										{result.metadata.author}
									</p>
								)}
							</div>
						)}
					</div>
				)}
			</div>

			<hr className="border-border" />

			{/* Library section */}
			<div className="flex flex-col gap-4">
				<div className="flex items-center gap-3">
					<Library className="size-6" />
					<h2 className="text-2xl font-bold">My Library</h2>
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
		</div>
	);
}
