import type { UploadResult } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";
import { ProcessFile } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/service";
import { createFileRoute } from "@tanstack/react-router";
import {
	Book,
	CheckCircle2,
	File,
	FileText,
	Loader2,
	Upload,
} from "lucide-react";
import { useCallback, useRef, useState } from "react";
import { Button } from "#/components/ui/button";

export const Route = createFileRoute("/upload")({ component: UploadPage });

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

function UploadPage() {
	const [file, setFile] = useState<File | null>(null);
	const [language, setLanguage] = useState("english");
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [result, setResult] = useState<UploadResult | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const [dragging, setDragging] = useState(false);

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
		setLoading(true);
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
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		} finally {
			setLoading(false);
		}
	};

	return (
		<div className="mx-auto flex max-w-3xl flex-col gap-6 p-6">
			<h1 className="text-2xl font-bold">Upload Document</h1>

			{/* Drop zone */}
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
					<label
						htmlFor="language-select"
						className="text-xs font-medium text-muted-foreground"
					>
						Language
					</label>
					<select
						id="language-select"
						value={language}
						onChange={(e) => setLanguage(e.target.value)}
						className="h-8 rounded-none border border-border bg-background px-2 text-xs outline-none focus:border-ring"
					>
						{LANGUAGES.map((l) => (
							<option key={l.value} value={l.value}>
								{l.label}
							</option>
						))}
					</select>
				</div>
				<Button onClick={handleProcess} disabled={!file || loading}>
					{loading ? (
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

			{/* Results */}
			{result && (
				<div className="flex flex-col gap-4">
					<div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
						<CheckCircle2 className="size-4" />
						Processed {result.sentenceCount} sentences from{" "}
						{result.totalChars.toLocaleString()} characters
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

					<div className="flex flex-col gap-1.5">
						{result.sentences.map((s) => (
							<div
								key={s.index}
								className="flex gap-3 border-l-2 border-muted-foreground/20 px-3 py-1.5 text-sm hover:border-primary/40"
							>
								<span className="mt-0.5 shrink-0 text-xs text-muted-foreground">
									{s.index}
								</span>
								<p>{s.text}</p>
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	);
}
