set dotenv-load := true

server := env_var("DEPLOY_SERVER")
path := env_var_or_default("DEPLOY_PATH", "/srv/spotify-saver")

deploy:
  go build -ldflags="-s -w" .
  ssh "{{ server }}" -- mkdir -p "{{ path }}/"
  rsync -P -e ssh "spotify" "config.json" "token.json" "{{ server }}:{{ path }}/"

logs:
  ssh "{{ server }}" -- less "{{ path }}/spotify.log"
