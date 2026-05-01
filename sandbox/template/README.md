# Shield Sandbox Project

A sample repo pre-configured for trying Shield.

## Layout

- `.shield`         — glob patterns of files to encrypt
- `.shieldignore`   — patterns to skip
- `secrets/`        — fake credentials
- `keys/`           — a fake PEM
- `config/`         — a fake admin config

## Try it

Encrypt the matched files in place:

    shield -e

Look at one of the files to see the `SHIELD[1.0]:` tag prefix:

    cat secrets/api-keys.txt

Decrypt them again:

    shield -d

Notice `secrets/public.txt` stays clear — `.shieldignore` skips it.

The vault password file lives at `~/.ssh/vault`.
