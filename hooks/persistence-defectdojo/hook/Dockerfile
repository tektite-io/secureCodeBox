# SPDX-FileCopyrightText: the secureCodeBox authors
#
# SPDX-License-Identifier: Apache-2.0

FROM gradle:jdk17 AS build
COPY . /home/gradle/src
WORKDIR /home/gradle/src
RUN ./gradlew build -x test

FROM gcr.io/distroless/java17-debian12:nonroot
COPY --from=build --chown=root:root --chmod=755 /home/gradle/src/build/libs/defectdojo-persistenceprovider-*.jar /app/defectdojo-persistenceprovider.jar
WORKDIR /app
# TLS Config works around an issue in OpenJDK... See: https://github.com/kubernetes-client/java/issues/854
ENTRYPOINT ["java", "-Djdk.tls.client.protocols=TLSv1.2", "-jar", "/app/defectdojo-persistenceprovider.jar"]
