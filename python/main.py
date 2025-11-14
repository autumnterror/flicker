# main.py
import os
import tempfile

from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import JSONResponse
from faster_whisper import WhisperModel

app = FastAPI(
    title="Whisper Transcription Service",
    description="Принимает аудио-файл и возвращает текст",
    version="1.0.0",
)

# ===== Настройки модели через переменные окружения =====
MODEL_SIZE = os.getenv("WHISPER_MODEL", "small")        # small / medium / large-v3 и т.д.
DEVICE = os.getenv("WHISPER_DEVICE", "cuda")            # "cuda" или "cpu"
COMPUTE_TYPE = os.getenv("WHISPER_COMPUTE_TYPE", "float16")  # "float16", "int8_float16" и т.п.

# ===== Инициализация модели (один раз на старт приложения) =====
print(f"Loading Whisper model '{MODEL_SIZE}' on device '{DEVICE}' with compute_type '{COMPUTE_TYPE}'...")
model = WhisperModel(
    MODEL_SIZE,
    device=DEVICE,
    compute_type=COMPUTE_TYPE,
)
print("Model loaded.")


@app.post("/transcribe")
async def transcribe(file: UploadFile = File(...)):
    """
    Принимает аудио-файл (mp3, wav, m4a, flac, ogg и т.д.) и возвращает распознанный текст.
    """
    if not file.content_type.startswith("audio/"):
        raise HTTPException(status_code=400, detail="Ожидался аудио-файл (content-type audio/*).")

    # Сохраняем во временный файл (faster-whisper удобно работает по пути к файлу)
    suffix = os.path.splitext(file.filename or "")[1] or ".tmp"
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            tmp_path = tmp.name
            content = await file.read()
            tmp.write(content)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Ошибка при сохранении файла: {e}")

    try:
        # language="ru" — форсируем русский; можно поставить None для автоопределения
        segments, info = model.transcribe(
            tmp_path,
            language="ru",
            beam_size=5,
            vad_filter=True,
        )

        parts = []
        for seg in segments:
            text = seg.text.strip()
            if text:
                parts.append(text)

        full_text = " ".join(parts)

        return JSONResponse(
            {
                "text": full_text,
                "language": info.language,
                "duration": info.duration,
                "language_probability": info.language_probability,
            }
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Ошибка при распознавании: {e}")
    finally:
        # Чистим временный файл
        try:
            os.remove(tmp_path)
        except Exception:
            pass


@app.get("/health")
def health():
    return {"status": "ok"}
