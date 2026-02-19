interface AudioPlaybackProps {
    audioUrl: string | null;
}

// Single concern: render the audio player when a recording is available
export default function AudioPlayback({ audioUrl }: AudioPlaybackProps) {
    if (!audioUrl) return null;

    return (
        <div className="playback-container">
            <audio controls src={audioUrl} />
        </div>
    );
}
