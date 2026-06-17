import { useCallback, useEffect, useRef, useState } from "react";
import { Link, createFileRoute, useNavigate } from "@tanstack/react-router";
import {
	ArrowLeft,
	BookOpen,
	Clock,
	Download,
	FileText,
	Globe,
	Loader2,
	Pause,
	Play,
	Square,
	Trash,
	User,
} from "lucide-react";
import { Events } from "@wailsio/runtime";

import { DeleteBook, GetBook } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/service";
import type { BookDetail } from "@bindings/github.com/vinewz/PageVoice/internal/services/textupload/models";
import {
	DownloadVoice,
	SplitBook,
	StartSynthesis,
	StopSynthesis,
	GetVoices,
	GetSynthesisStatus,
	GetSentences,
	GetAudio,
	GetGeneratedChunks,
} from "@bindings/github.com/vinewz/PageVoice/internal/services/tts/service";
import type { VoiceInfo, SynthesisProgress } from "@bindings/github.com/vinewz/PageVoice/internal/services/tts/models";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

export const Route = createFileRoute("/books/$dirName")({
	component: BookPage,
});

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

interface SentenceData {
	index: number;
	text: string;
	chunk: number;
}

interface SentencesFile {
	language: string;
	chunkLength: number;
	totalChunks: number;
	sentences: SentenceData[];
}

function BookPage() {
	const { dirName } = Route.useParams();
	const navigate = useNavigate();
	const [book, setBook] = useState<BookDetail | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [voices, setVoices] = useState<VoiceInfo[]>([]);
	const [sentencesFile, setSentencesFile] = useState<SentencesFile | null>(null);
	const [progress, setProgress] = useState<SynthesisProgress | null>(null);
	const [language, setLanguage] = useState("english");
	const [selectedVoice, setSelectedVoice] = useState("");
	const [running, setRunning] = useState(false);
	const [currentChunk, setCurrentChunk] = useState(0);
	const [currentAudioChunk, setCurrentAudioChunk] = useState<number | null>(null);
	const [currentAudioSrc, setCurrentAudioSrc] = useState("");
	const [generatedChunks, setGeneratedChunks] = useState<Set<number>>(new Set());
	const audioRef = useRef<HTMLAudioElement>(null);

	useEffect(() => {
		return () => {
			if (currentAudioSrc) URL.revokeObjectURL(currentAudioSrc);
		};
	}, [currentAudioSrc]);

	useEffect(() => {
		if (currentAudioSrc && audioRef.current) {
			audioRef.current.play();
		}
	}, [currentAudioSrc]);

	const loadData = useCallback(async () => {
		try {
			const [b, v, p] = await Promise.all([
				GetBook(dirName),
				GetVoices(),
				GetSynthesisStatus(dirName),
			]);
			setBook(b);
			setVoices(v);
			if (b) setLanguage(b.metadata.language || "english");
			if (p) setProgress(p);

			if (p && p.status === "running") {
				setRunning(true);
				setCurrentChunk(p.currentChunk);
			}

			try {
				const s = await GetSentences(dirName);
				if (s) {
					setSentencesFile(JSON.parse(s));
				}
			} catch {
				// no sentences yet
			}

			const chunks = await GetGeneratedChunks(dirName);
			if (chunks) {
				setGeneratedChunks(new Set(chunks));
			}
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		} finally {
			setLoading(false);
		}
	}, [dirName]);

	useEffect(() => {
		loadData();
	}, [loadData]);

	useEffect(() => {
		const offStart = Events.On("tts:chunk-start", (event: unknown) => {
			const data = (event as { data: { chunk: number } }).data;
			setCurrentChunk(data.chunk);
		});

		const offComplete = Events.On("tts:chunk-complete", async (event: unknown) => {
			const data = (event as { data: { chunk: number; total: number } }).data;
			setCurrentChunk(data.chunk + 1);
			setProgress((prev) => prev ? { ...prev, currentChunk: data.chunk + 1 } : prev);
			setGeneratedChunks((prev) => new Set(prev).add(data.chunk + 1));
			if (data.chunk + 1 >= data.total) {
				setRunning(false);
				loadData();
			}
		});

		const offError = Events.On("tts:chunk-error", () => {
			setRunning(false);
			loadData();
		});

		return () => {
			offStart();
			offComplete();
			offError();
		};
	}, [loadData]);

	const handleLoadBook = async () => {
		setLoading(true);
		try {
			await SplitBook(dirName, language);
			const s = await GetSentences(dirName);
			if (s) {
				setSentencesFile(JSON.parse(s));
			}
			loadData();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		} finally {
			setLoading(false);
		}
	};

	const handleStart = async () => {
		if (!selectedVoice) return;
		setRunning(true);
		setError(null);
		try {
			await StartSynthesis(dirName, selectedVoice);
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
			setRunning(false);
		}
	};

	const handleStop = async () => {
		try {
			await StopSynthesis();
			setRunning(false);
			loadData();
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		}
	};

	const handlePlayChunk = async (chunkIdx: number) => {
		if (currentAudioChunk === chunkIdx) {
			if (audioRef.current?.paused) {
				audioRef.current.play();
			} else {
				audioRef.current?.pause();
			}
			return;
		}

		try {
			const b64 = await GetAudio(dirName, chunkIdx);
			const binary = atob(b64);
			const bytes = new Uint8Array(binary.length);
			for (let i = 0; i < binary.length; i++) {
				bytes[i] = binary.charCodeAt(i) & 0xff;
			}
			const blob = new Blob([bytes], { type: "audio/wav" });
			const url = URL.createObjectURL(blob);

			if (currentAudioSrc) URL.revokeObjectURL(currentAudioSrc);
			setCurrentAudioSrc(url);
			setCurrentAudioChunk(chunkIdx);
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		}
	};

	const handlePlayAll = async () => {
		if (!sentencesFile) return;
		const sorted = Array.from(generatedChunks).sort((a, b) => a - b);
		if (sorted.length === 0) return;
		await handlePlayChunk(sorted[0]);
	};

	const handleDelete = async () => {
		if (!window.confirm("Delete this book and all its audio?")) return;
		try {
			await DeleteBook(dirName);
			navigate({ to: "/" });
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		}
	};

	const handleDownloadVoice = async (name: string) => {
		try {
			await DownloadVoice(name);
			setVoices((prev) =>
				prev.map((v) => (v.name === name ? { ...v, downloaded: true } : v)),
			);
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(msg);
		}
	};

	if (loading) {
		return (
			<div className="mx-auto flex max-w-2xl p-6">
				<p className="text-sm text-muted-foreground">Loading...</p>
			</div>
		);
	}

	if (error && !book) {
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

	const status = progress?.status || book?.state.status || "pending";
	const totalChunks = sentencesFile?.totalChunks || progress?.totalChunks || 0;
	const pct = totalChunks > 0 ? Math.round((currentChunk / totalChunks) * 100) : 0;
	return (
		<div className="mx-auto flex max-w-2xl flex-col gap-6 p-6">
			<Link
				to="/"
				className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
			>
				<ArrowLeft className="size-4" />
				Back to library
			</Link>

			{/* Book metadata */}
			<div className="flex items-center gap-3">
				<BookOpen className="size-6" />
				<h1 className="text-2xl font-bold">{book?.metadata.title}</h1>
				<button
					type="button"
					className="ml-auto flex items-center gap-1 text-xs text-muted-foreground hover:text-destructive"
					onClick={handleDelete}
				>
					<Trash className="size-4" />
				</button>
			</div>

			{error && (
				<div className="border border-destructive/50 bg-destructive/10 px-4 py-2 text-xs text-destructive">
					{error}
				</div>
			)}

			<div className="flex flex-col gap-2 border border-border bg-muted/30 px-4 py-3 text-xs">
				{book?.metadata.author && (
					<div className="flex items-center gap-2">
						<User className="size-3.5 text-muted-foreground" />
						<span className="text-muted-foreground">Author:</span>
						<span>{book.metadata.author}</span>
					</div>
				)}
				<div className="flex items-center gap-2">
					<Globe className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Language:</span>
					<span>{book?.metadata.language}</span>
				</div>
				<div className="flex items-center gap-2">
					<FileText className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Source:</span>
					<span>{book?.metadata.sourceFile}</span>
				</div>
				<div className="flex items-center gap-2">
					<Clock className="size-3.5 text-muted-foreground" />
					<span className="text-muted-foreground">Imported:</span>
					<span>{book?.metadata.importedAt ? new Date(book.metadata.importedAt).toLocaleString() : ""}</span>
				</div>
				<div className="mt-1 flex items-center gap-2 border-t border-border pt-2">
					<span className="text-muted-foreground">Status:</span>
					<span className={status === "completed" ? "text-green-600 dark:text-green-400" : ""}>{status}</span>
				</div>
			</div>

			{/* Table of Contents */}
			{book?.metadata.toc && book.metadata.toc.length > 0 && (
				<div className="flex flex-col gap-1.5">
					<h2 className="text-sm font-semibold text-muted-foreground">
						Table of Contents ({book.metadata.toc.length})
					</h2>
					<div className="max-h-[300px] overflow-y-auto border border-border">
						{book.metadata.toc.map((entry, i) => (
							<div
								key={entry.title + entry.depth}
								className={`flex items-center gap-2 px-3 py-1 text-xs transition-colors hover:bg-muted/50 ${
									i % 2 === 0 ? "bg-muted/20" : ""
								}`}
								style={{ paddingLeft: `${12 + entry.depth * 16}px` }}
							>
								<span className="text-muted-foreground">{entry.title}</span>
							</div>
						))}
					</div>
				</div>
			)}

			{/* Load Book (split sentences) */}
			{!sentencesFile && (
				<div className="flex items-end gap-3">
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
										<SelectItem key={l.value} value={l.value}>{l.label}</SelectItem>
									))}
								</SelectGroup>
							</SelectContent>
						</Select>
					</div>
					<Button onClick={handleLoadBook}>
						{loading ? <Loader2 className="size-4 animate-spin" /> : null}
						Load Book
					</Button>
				</div>
			)}

			{/* Voice selector + Synthesis controls */}
			{sentencesFile && status !== "completed" && (
				<div className="flex flex-wrap items-end gap-3">
					<div className="flex flex-col gap-1.5">
						<span className="text-xs font-medium text-muted-foreground">
							Voice
						</span>
						<Select value={selectedVoice} onValueChange={(v) => setSelectedVoice(v ?? "")}>
							<SelectTrigger className="w-[240px]">
								<SelectValue placeholder="Select a voice..." />
							</SelectTrigger>
							<SelectContent>
								<SelectGroup>
									{voices.map((v) => (
										<SelectItem key={v.name} value={v.name}>
											{v.name} ({v.quality}){v.downloaded ? "" : " — not downloaded"}
										</SelectItem>
									))}
								</SelectGroup>
							</SelectContent>
						</Select>
					</div>

					{selectedVoice && !voices.find((v) => v.name === selectedVoice)?.downloaded && (
						<Button
							variant="outline"
							size="xs"
							onClick={() => handleDownloadVoice(selectedVoice)}
						>
							<Download className="size-3" />
							Download
						</Button>
					)}

					{status === "running" || running ? (
						<Button variant="destructive" onClick={handleStop} className="gap-1.5">
							<Square className="size-3.5" />
							Stop
						</Button>
					) : (
						<Button onClick={handleStart} disabled={!selectedVoice} className="gap-1.5">
							<Play className="size-3.5" />
							Start Synthesis
						</Button>
					)}
				</div>
			)}

			{/* Progress bar */}
			{running && totalChunks > 0 && (
				<div className="flex flex-col gap-1.5">
					<div className="flex items-center justify-between text-xs text-muted-foreground">
						<span>Chunk {currentChunk + 1} of {totalChunks}</span>
						<span>{pct}%</span>
					</div>
					<div className="h-1.5 w-full bg-muted">
						<div
							className="h-full bg-primary transition-all duration-300"
							style={{ width: `${pct}%` }}
						/>
					</div>
				</div>
			)}

			{/* Completed header */}
			{status === "completed" && (
				<div className="flex items-center gap-2 border border-green-500/30 bg-green-500/10 px-4 py-2 text-xs text-green-700 dark:text-green-300">
					<span>Synthesis complete. All {totalChunks} chunks generated.</span>
					{totalChunks > 0 && (
						<Button variant="outline" size="xs" onClick={handlePlayAll}>
							<Play className="size-3" />
							Play All
						</Button>
					)}
				</div>
			)}

			{/* Sentences display */}
			{sentencesFile && (
				<div className="flex flex-col gap-1.5">
					<h2 className="text-sm font-semibold text-muted-foreground">
						Sentences ({sentencesFile.sentences.length})
					</h2>
					<div className="max-h-[500px] overflow-y-auto border border-border">
						{sentencesFile.sentences.map((s) => {
							const faded = generatedChunks.has(s.chunk + 1);
							return (
								<div
									key={s.index}
									className={`px-3 py-1.5 text-sm leading-relaxed transition-colors hover:bg-muted/50 ${
										faded ? "text-muted-foreground/50" : ""
									} ${s.index % 2 === 0 ? "bg-muted/20" : ""}`}
								>
									{s.text}
								</div>
							);
						})}
					</div>
				</div>
			)}

			{/* Generated chunks audio list */}
			{generatedChunks.size > 0 && (
				<div className="flex flex-col gap-1.5">
					<h2 className="text-sm font-semibold text-muted-foreground">
						Generated Chunks ({generatedChunks.size})
					</h2>

					{currentAudioChunk !== null && currentAudioSrc && (
						<audio
							ref={audioRef}
							controls
							src={currentAudioSrc}
							onEnded={() => setCurrentAudioChunk(null)}
							className="w-full"
						/>
					)}

					<div className="max-h-[500px] overflow-y-auto border border-border">
						{Array.from(generatedChunks).sort((a, b) => a - b).map((chunkIdx, i) => (
							<div
								key={chunkIdx}
								className={`flex items-center gap-2 px-3 py-1.5 text-sm transition-colors hover:bg-muted/50 ${
									i % 2 === 0 ? "bg-muted/20" : ""
								}`}
							>
								<button
									type="button"
									className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
									onClick={() => handlePlayChunk(chunkIdx)}
								>
									{currentAudioChunk === chunkIdx ? (
										<Pause className="size-3.5" />
									) : (
										<Play className="size-3.5" />
									)}
								</button>
								<span className="font-mono text-xs text-muted-foreground">
									Chunk {String(chunkIdx).padStart(3, "0")}
								</span>
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	);
}
