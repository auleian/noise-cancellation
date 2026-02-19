import micIcon from "../assets/mic-1.svg"

interface MicButtonProps {
    active: boolean;
    onClick: () => void;
}

// Single concern: render the clickable microphone button
export default function MicButton({ active, onClick }: MicButtonProps) {
    return (
        <div
            className={`mic-border ${active ? "mic-border-active" : ""}`}
            onClick={onClick}
        >
            <img
                src={micIcon}
                alt="microphone"
                className={`mic-icon ${active ? "mic-active" : ""}`}
            />
        </div>
    );
}
