# Copyright 2015 The Kubernetes Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# BITE-1345 image source previously gcr.io/google-containers/nginx:latest

FROM nginx:1.12.0
RUN apt update && apt install -y openssl procps
COPY controller /
COPY default.conf /etc/nginx/nginx.conf
COPY /run.sh /usr/bin/run.sh
RUN chmod +x /usr/bin/run.sh
COPY nginx.conf.tmpl /etc/nginx/
COPY healthz.tmpl /etc/nginx/
# cops-165
COPY error_page.tmpl /etc/nginx/
# To overide the default nginx page
COPY index.html /usr/share/nginx/html/index.html
COPY pearson_logo.png /usr/share/nginx/html/
COPY favicon.ico /usr/share/nginx/html/
# BITE-1345
COPY dhparam.pem /etc/nginx/certs/

CMD ["/usr/bin/run.sh"]
