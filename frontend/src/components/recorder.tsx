import { useRecorder } from "../hooks/useRecorder"
import WaveAnimation from "./WaveAnimation"
import MicButton from "./MicButton"
import AudioPlayback from "./AudioPlayback"
import "./recorder.css"

export default function Recorder() {
    const {
        active,
        audioUrl,
        cleanedAudioUrl,
        processing,
        toggleRecording,
        deleteRecording,
        sendRecording,
    } = useRecorder();

    return (
        <div className="recorder-container">
            <WaveAnimation active={active} />
            <MicButton active={active} onClick={toggleRecording} />
            <AudioPlayback
                audioUrl={audioUrl}
                cleanedAudioUrl={cleanedAudioUrl}
                processing={processing}
                onDelete={deleteRecording}
                onSend={sendRecording}
            />
        </div>
    );
}
