FROM node:24-alpine AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME/bin:$PATH"
RUN corepack enable
WORKDIR /app

FROM base AS prod-deps
COPY package.json pnpm-lock.yaml /app/
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile

FROM base AS build
ARG PUBLIC_SERVER_URL
ARG PUBLIC_WS_BASE
ENV PUBLIC_SERVER_URL=$PUBLIC_SERVER_URL
ENV PUBLIC_WS_BASE=$PUBLIC_WS_BASE
COPY . /app/
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM base
COPY package.json /app/
COPY --from=prod-deps /app/node_modules /app/node_modules
COPY --from=build /app/build /app/build
USER node
ENV NODE_ENV=production
ENV PORT=3001
EXPOSE 3001
CMD ["node", "build"]
