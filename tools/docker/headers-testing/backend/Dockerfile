# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM python:latest

COPY ./headertest.py /
RUN chmod +x /headertest.py

EXPOSE 80

CMD ["python", "/headertest.py"]
