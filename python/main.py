# main.py
import os
import tempfile
from io import BytesIO
from pathlib import Path
from typing import List, Dict, Any

from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import JSONResponse
from pydub import AudioSegment
from deepgram import DeepgramClient

DEFAULT_EXPORT_FORMAT = "flac"

APP_TITLE = "Deepgram HTTP Transcriber"

# Настройки через переменные окружения
DEEPGRAM_API_KEY = os.getenv("DEEPGRAM_API_KEY", "")
if not DEEPGRAM_API_KEY:
    raise RuntimeError("DEEPGRAM_API_KEY не задан. Установи его в переменных окружения.")

DEEPGRAM_MODEL = os.getenv("DEEPGRAM_MODEL", "nova-2-general")  # подставь нужную модель Deepgram
LANGUAGE = os.getenv("DEEPGRAM_LANGUAGE", "ru")                  # язык ("ru", "en", или None/"" для авто)
SEGMENT_MINUTES = int(os.getenv("SEGMENT_MINUTES", "10"))        # длина сегмента в минутах
SMART_FORMAT = os.getenv("SMART_FORMAT", "true").lower() == "true"

# Инициализируем Deepgram-клиент один раз
dg_client = DeepgramClient(api_key=DEEPGRAM_API_KEY)

app = FastAPI(
    title=APP_TITLE,
    description="Принимает аудио/видео файл, нарезает на сегменты и расшифровывает через Deepgram API.",
    version="1.0.0",
)


def load_audio_from_temp(path: Path) -> AudioSegment:
    """
    Упрощённый вариант: грузим файл через pydub/ffmpeg.
    Deepgram GUI-скрипт делал танцы с Unicode-путями, нам пока хватит этого.
    """
    try:
        return AudioSegment.from_file(path)
    except Exception as e:
        raise RuntimeError(f"Не удалось открыть файл через FFmpeg/pydub: {e}") from e


@app.post("/transcribe")
async def transcribe(file: UploadFile = File(...)):
    """
    Принимает аудио/видео файл (mp3, wav, m4a, mp4 и т.п.) и возвращает:
    - полный текст
    - список сегментов (start/end в секундах + текст)
    """
    if not file.filename:
        raise HTTPException(status_code=400, detail="Файл не передан")

    # Сохраняем загрузку во временный файл (pydub любит пути)
    suffix = Path(file.filename).suffix or ".tmp"

    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            tmp_path = Path(tmp.name)
            content = await file.read()
            tmp.write(content)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Ошибка при сохранении временного файла: {e}")

    try:
        # 1) Загружаем аудио/видео целиком
        full_audio = load_audio_from_temp(tmp_path)
        total_ms = len(full_audio)

        if total_ms == 0:
            raise HTTPException(status_code=400, detail="Похоже, файл пустой или не распознан как аудио.")

        seg_ms = max(1, SEGMENT_MINUTES) * 60 * 1000
        total_segments = (total_ms + seg_ms - 1) // seg_ms

        # 2) Собираем полный текст и данные по сегментам
        full_transcript_parts: List[str] = []
        segments_data: List[Dict[str, Any]] = []

        for i in range(0, total_ms, seg_ms):
            seg_idx = i // seg_ms + 1
            start_time = i
            end_time = min(i + seg_ms, total_ms)
            segment = full_audio[start_time:end_time]

            # Экспортируем кусок в FLAC в память
            buf = BytesIO()
            segment.export(buf, format=DEFAULT_EXPORT_FORMAT)
            buf.seek(0)

            try:
                # ВАЖНО: здесь тот же вызов, что и в dp.py,
                # только без UI и очередей — прямой запрос к SDK.
                response = dg_client.listen.v1.media.transcribe_file(
                    request=buf.read(),
                    model=DEEPGRAM_MODEL,
                    smart_format=SMART_FORMAT,
                    language=LANGUAGE or None,
                )

                text = response.results.channels[0].alternatives[0].transcript
                text = (text or "").strip()

                full_transcript_parts.append(text)
                segments_data.append(
                    {
                        "segment_index": seg_idx,
                        "start_sec": start_time // 1000,
                        "end_sec": end_time // 1000,
                        "text": text,
                    }
                )
            except Exception as e:
                # В реальном сервисе сюда можно добавить логгер / отправку в Sentry
                raise HTTPException(
                    status_code=502,
                    detail=f"Ошибка Deepgram при обработке сегмента {seg_idx}: {e}",
                )

        transcript = (" ".join(full_transcript_parts)).strip()

        result = {
            "filename": file.filename,
            "duration_seconds": total_ms / 1000.0,
            "segment_minutes": SEGMENT_MINUTES,
            "segments_count": total_segments,
            "language": LANGUAGE,
            "model": DEEPGRAM_MODEL,
            "text": transcript,
            "segments": segments_data,
        }

        return JSONResponse(result)

    finally:
        try:
            tmp_path.unlink(missing_ok=True)
        except Exception:
            pass


@app.get("/health")
def health():
    return {"status": "ok"}
