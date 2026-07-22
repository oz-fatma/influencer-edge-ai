# MLC wheels are published for linux/amd64 (manylinux), not linux/arm64.
# On Apple Silicon: docker build --platform linux/amd64 -t influencer-llm-service .
FROM python:3.11-slim AS builder

WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive \
  LD_LIBRARY_PATH=/usr/local/lib/python3.11/site-packages/tvm_ffi/lib:/usr/local/lib/python3.11/site-packages/tvm/lib:/usr/local/lib/python3.11/site-packages/mlc_llm/lib

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    git-lfs \
    libgomp1 \
    libx11-6 \
    build-essential \
  && git lfs install \
  && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --upgrade pip \
  && pip install --no-cache-dir --pre \
    -f https://mlc.ai/wheels \
    --only-binary=:all: \
    --no-deps \
    "mlc-ai-cpu==0.20.0" \
    "mlc-llm-cpu==0.20.0.dev0" \
  && pip install --no-cache-dir "apache-tvm-ffi==0.1.10" \
  && pip install --no-cache-dir \
    --index-url https://download.pytorch.org/whl/cpu \
    --extra-index-url https://pypi.org/simple \
    torch \
  && pip install --no-cache-dir \
    cloudpickle \
    "ml_dtypes>=0.5.1" \
    numpy \
    packaging \
    psutil \
    pytest \
    scipy \
    tornado \
    typing_extensions \
    datasets \
    fastapi \
    openai \
    pandas \
    prompt_toolkit \
    requests \
    safetensors \
    sentencepiece \
    shortuuid \
    tiktoken \
    tqdm \
    transformers \
    uvicorn \
  && rm -f /usr/local/lib/python3.11/site-packages/mlc_llm/lib/libtvm_runtime.so \
  && ln -sf ../../tvm/lib/libtvm_runtime.so \
    /usr/local/lib/python3.11/site-packages/mlc_llm/lib/libtvm_runtime.so

# Download Gemma weights + compile CPU model library (skips runtime JIT).
RUN git clone --depth 1 https://huggingface.co/mlc-ai/gemma-2b-it-q4f16_1-MLC /app/model \
  && git -C /app/model lfs pull \
  && python -m mlc_llm compile /app/model \
    --device cpu \
    --quantization q4f16_1 \
    --output /app/model/gemma-cpu.so

FROM python:3.11-slim

WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive \
  LD_LIBRARY_PATH=/usr/local/lib/python3.11/site-packages/tvm_ffi/lib:/usr/local/lib/python3.11/site-packages/tvm/lib:/usr/local/lib/python3.11/site-packages/mlc_llm/lib \
  MODEL=/app/model \
  MODEL_LIB=/app/model/gemma-cpu.so

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    libgomp1 \
    libx11-6 \
  && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --upgrade pip \
  && pip install --no-cache-dir --pre \
    -f https://mlc.ai/wheels \
    --only-binary=:all: \
    --no-deps \
    "mlc-ai-cpu==0.20.0" \
    "mlc-llm-cpu==0.20.0.dev0" \
  && pip install --no-cache-dir "apache-tvm-ffi==0.1.10" \
  && pip install --no-cache-dir \
    --index-url https://download.pytorch.org/whl/cpu \
    --extra-index-url https://pypi.org/simple \
    torch \
  && pip install --no-cache-dir \
    cloudpickle \
    "ml_dtypes>=0.5.1" \
    numpy \
    packaging \
    psutil \
    pytest \
    scipy \
    tornado \
    typing_extensions \
    datasets \
    fastapi \
    openai \
    pandas \
    prompt_toolkit \
    requests \
    safetensors \
    sentencepiece \
    shortuuid \
    tiktoken \
    tqdm \
    transformers \
    uvicorn \
  && rm -f /usr/local/lib/python3.11/site-packages/mlc_llm/lib/libtvm_runtime.so \
  && ln -sf ../../tvm/lib/libtvm_runtime.so \
    /usr/local/lib/python3.11/site-packages/mlc_llm/lib/libtvm_runtime.so

# Lazy-import embedding path (optional; needs full TVM compiler).
RUN python - <<'PY'
from pathlib import Path

serve_init = Path("/usr/local/lib/python3.11/site-packages/mlc_llm/serve/__init__.py")
text = serve_init.read_text()
old = "from .embedding_engine import AsyncEmbeddingEngine\n"
if old in text:
    serve_init.write_text(text.replace(old, "# patched: lazy embedding import\n", 1))

iface = Path("/usr/local/lib/python3.11/site-packages/mlc_llm/interface/serve.py")
text = iface.read_text()
text = text.replace("from mlc_llm.serve.embedding_engine import AsyncEmbeddingEngine\n", "")
old = """    if embedding_model is not None:
        if embedding_model_lib is None:
"""
new = """    if embedding_model is not None:
        from mlc_llm.serve.embedding_engine import AsyncEmbeddingEngine
        if embedding_model_lib is None:
"""
if old in text:
    iface.write_text(text.replace(old, new, 1))
PY

COPY --from=builder /app/model /app/model

EXPOSE 8000

CMD ["sh", "-c", "python3 -m mlc_llm serve \"$MODEL\" --model-lib \"$MODEL_LIB\" --host 0.0.0.0 --port 8000 --device cpu"]
