interface AudioPlaybackProps {
    audioUrl: string | null;
    cleanedAudioUrl: string | null;
    processing: boolean;
    onDelete: () => void;
    onSend: () => void;
}

export default function AudioPlayback({
    audioUrl,
    cleanedAudioUrl,
    processing,
    onDelete,
    onSend,
}: AudioPlaybackProps) {
    if (!audioUrl) return null;

    // Both original and cleaned available — show side by side for comparison.
    if (cleanedAudioUrl) {
        return (
            <div className="playback-container">
                <div className="audio-bubble audio-bubble-original">
                    <span className="original-label">Original</span>
                    <audio
                        key={audioUrl}
                        controls
                        src={audioUrl}
                        style={{ display: "block", width: "100%", minHeight: 54 }}
                    />
                </div>
                <div className="audio-bubble audio-bubble-cleaned">
                    <span className="cleaned-label">Cleaned</span>
                    <audio
                        key={cleanedAudioUrl}
                        controls
                        src={cleanedAudioUrl}
                        style={{ display: "block", width: "100%", minHeight: 54 }}
                    />
                    <a
                        href={cleanedAudioUrl}
                        download="cleaned.wav"
                        className="download-link"
                    >
                        Download
                    </a>
                </div>
            </div>
        );
    }

    // Not yet sent — show send/delete controls.
    return (
        <div className="playback-container">
            <div className="audio-bubble">
                <audio controls src={audioUrl} />
                <div className="audio-actions">
                    <button
                        className="btn-delete"
                        onClick={onDelete}
                        disabled={processing}
                        title="Delete"
                    >
                        <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM8 9h8v10H8V9zm7.5-5l-1-1h-5l-1 1H5v2h14V4h-3.5z" />
                        </svg>
                    </button>
                    <button
                        className="btn-send"
                        onClick={onSend}
                        disabled={processing}
                        title="Send for noise cancellation"
                    >
                        {processing ? (
                            <span className="spinner" />
                        ) : (
                            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
                                <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
                            </svg>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
}
