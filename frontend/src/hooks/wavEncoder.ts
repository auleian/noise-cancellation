// Encode raw PCM Float32Array samples into a 16-bit mono WAV Blob.
export function encodeWAV(samples: Float32Array, sampleRate: number): Blob {
    const buffer = new ArrayBuffer(44 + samples.length * 2);
    const view = new DataView(buffer);

    // RIFF header
    writeString(view, 0, "RIFF");
    view.setUint32(4, 36 + samples.length * 2, true);
    writeString(view, 8, "WAVE");

    // fmt chunk
    writeString(view, 12, "fmt ");
    view.setUint32(16, 16, true);              // chunk size
    view.setUint16(20, 1, true);               // PCM format
    view.setUint16(22, 1, true);               // mono
    view.setUint32(24, sampleRate, true);       // sample rate
    view.setUint32(28, sampleRate * 2, true);   // byte rate (sampleRate * channels * bytesPerSample)
    view.setUint16(32, 2, true);               // block align
    view.setUint16(34, 16, true);              // bits per sample

    // data chunk
    writeString(view, 36, "data");
    view.setUint32(40, samples.length * 2, true);

    let offset = 44;
    for (let i = 0; i < samples.length; i++) {
        const s = Math.max(-1, Math.min(1, samples[i]));
        view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7fff, true);
        offset += 2;
    }

    return new Blob([buffer], { type: "audio/wav" });
}

function writeString(view: DataView, offset: number, str: string): void {
    for (let i = 0; i < str.length; i++) {
        view.setUint8(offset + i, str.charCodeAt(i));
    }
}

// Convert a WebM blob to a mono WAV blob using the Web Audio API.
export async function webmToWav(webmBlob: Blob): Promise<Blob> {
    const arrayBuffer = await webmBlob.arrayBuffer();
    const audioCtx = new AudioContext();

    try {
        const audioBuffer = await audioCtx.decodeAudioData(arrayBuffer);

        // Mix down to mono by averaging all channels.
        const length = audioBuffer.length;
        const mono = new Float32Array(length);
        const numChannels = audioBuffer.numberOfChannels;

        for (let ch = 0; ch < numChannels; ch++) {
            const channelData = audioBuffer.getChannelData(ch);
            for (let i = 0; i < length; i++) {
                mono[i] += channelData[i];
            }
        }

        if (numChannels > 1) {
            for (let i = 0; i < length; i++) {
                mono[i] /= numChannels;
            }
        }

        return encodeWAV(mono, audioBuffer.sampleRate);
    } finally {
        await audioCtx.close();
    }
}
