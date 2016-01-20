# Copyright 2016 Ciena Corporation
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# you may obtain a copy of the License at
#
#    http://www.apache.org/license/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, sofware
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM golang

ADD ./src/configserver /go/src/configserver
ADD ./app/configserv.go /go/src/app/configserv.go

RUN go install app

ENTRYPOINT /go/bin/app

VOLUME [ "/configuration" ]
EXPOSE 4321

