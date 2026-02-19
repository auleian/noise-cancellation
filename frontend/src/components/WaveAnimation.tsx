// Single concern: render pulsing wave rings when recording is active
export default function WaveAnimation({ active }: { active: boolean }) {
    if (!active) return null;

    return (
        <>
            <div className="wave-ring wave-ring-1" />
            <div className="wave-ring wave-ring-2" />
            <div className="wave-ring wave-ring-3" />
        </>
    );
}
