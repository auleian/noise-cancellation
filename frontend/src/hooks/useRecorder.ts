import { useState, useRef, useCallback } from "react";
import { webmToWav } from "./wavEncoder";

const BACKEND_URL = "http://localhost:8080";

export function useRecorder() {
    const [active, setActive] = useState(false);
    const [audioUrl, setAudioUrl] = useState<string | null>(null);
    const [cleanedAudioUrl, setCleanedAudioUrl] = useState<string | null>(null);
    const [processing, setProcessing] = useState(false);

    const mediaRecorderRef = useRef<MediaRecorder | null>(null);
    const chunksRef = useRef<Blob[]>([]);
    const webmBlobRef = useRef<Blob | null>(null);

    const revokeUrls = useCallback(() => {
        if (audioUrl) URL.revokeObjectURL(audioUrl);
        if (cleanedAudioUrl) URL.revokeObjectURL(cleanedAudioUrl);
    }, [audioUrl, cleanedAudioUrl]);

    const clearPreviousRecording = useCallback(() => {
        revokeUrls();
        setAudioUrl(null);
        setCleanedAudioUrl(null);
        webmBlobRef.current = null;
    }, [revokeUrls]);

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
            webmBlobRef.current = blob;
            setAudioUrl(URL.createObjectURL(blob));
            stream.getTracks().forEach((track) => track.stop());
        };

        mediaRecorder.start();
        setActive(true);
    }, [clearPreviousRecording]);

    const stopRecording = useCallback(() => {
        mediaRecorderRef.current?.stop();
        setActive(false);
    }, []);

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

    // Discard the current recording.
    const deleteRecording = useCallback(() => {
        clearPreviousRecording();
    }, [clearPreviousRecording]);

    // Convert WebM to WAV, upload to backend, receive cleaned audio.
    const sendRecording = useCallback(async () => {
        if (!webmBlobRef.current) return;

        setProcessing(true);
        try {
            // Convert WebM → WAV.
            const wavBlob = await webmToWav(webmBlobRef.current);

            // Upload WAV to backend.
            const formData = new FormData();
            formData.append("file", wavBlob, "recording.wav");

            const response = await fetch(`${BACKEND_URL}/denoise`, {
                method: "POST",
                body: formData,
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(`Server error: ${text}`);
            }

            // Explicitly set blob type to audio/wav so the browser can play it.
            const arrayBuffer = await response.arrayBuffer();
            const cleanedBlob = new Blob([arrayBuffer], { type: "audio/wav" });

            // Set the cleaned audio for playback (keep original audioUrl for comparison).
            setCleanedAudioUrl(URL.createObjectURL(cleanedBlob));
        } catch (err) {
            console.error("Failed to process audio:", err);
        } finally {
            setProcessing(false);
        }
    }, [audioUrl]);

    return {
        active,
        audioUrl,
        cleanedAudioUrl,
        processing,
        toggleRecording,
        deleteRecording,
        sendRecording,
    };
}
