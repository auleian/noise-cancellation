import { useState, useRef, useCallback } from "react";

// Custom hook: single concern — manage audio recording lifecycle
export function useRecorder() {
    const [active, setActive] = useState(false);
    const [audioUrl, setAudioUrl] = useState<string | null>(null);

    const mediaRecorderRef = useRef<MediaRecorder | null>(null);
    const chunksRef = useRef<Blob[]>([]);

    // Single concern: clean up a previous recording's object URL
    const clearPreviousRecording = useCallback(() => {
        if (audioUrl) {
            URL.revokeObjectURL(audioUrl);
            setAudioUrl(null);
        }
    }, [audioUrl]);

    // Single concern: request mic access and start the MediaRecorder
    const startRecording = useCallback(async () => {
        clearPreviousRecording();

        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        const mediaRecorder = new MediaRecorder(stream);
        mediaRecorderRef.current = mediaRecorder;
        chunksRef.current = [];

        mediaRecorder.ondataavailable = (e) => {
            if (e.data.size > 0) {
                chunksRef.current.push(e.data);
            }
        };

        mediaRecorder.onstop = () => {
            const blob = new Blob(chunksRef.current, { type: "audio/webm" });
            setAudioUrl(URL.createObjectURL(blob));
            stream.getTracks().forEach((track) => track.stop());
        };

        mediaRecorder.start();
        setActive(true);
    }, [clearPreviousRecording]);

    // Single concern: stop the MediaRecorder
    const stopRecording = useCallback(() => {
        mediaRecorderRef.current?.stop();
        setActive(false);
    }, []);

    // Single concern: toggle between start and stop
    const toggleRecording = useCallback(async () => {
        if (!active) {
            try {
                await startRecording();
            } catch (err) {
                console.error("Microphone access denied:", err);
            }
        } else {
            stopRecording();
        }
    }, [active, startRecording, stopRecording]);

    return { active, audioUrl, toggleRecording };
}
