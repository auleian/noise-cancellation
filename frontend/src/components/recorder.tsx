import { useRecorder } from "../hooks/useRecorder"
import WaveAnimation from "./WaveAnimation"
import MicButton from "./MicButton"
import AudioPlayback from "./AudioPlayback"
import "./recorder.css"

// Single concern: compose recording sub-components and pass state down
export default function Recorder() {
    const { active, audioUrl, toggleRecording } = useRecorder();

    return (
        <div className="recorder-container">
            <WaveAnimation active={active} />
            <MicButton active={active} onClick={toggleRecording} />
            <AudioPlayback audioUrl={audioUrl} />
        </div>
    );
}
