# Mozeidon addon source-code

The source code is written in ``typescript`` and built with ``webpack``.
As specified in the ``package.json``, the ``node`` version should be >= 20 and the ``npm`` version should be >= 10.

The Chrome add-on shares its TypeScript source with the Firefox add-on. Build it from the repo root with:

```bash
just build-chrome
```

This syncs ``src/`` from ``firefox-addon/`` and produces ``dist/background.js``.

Note :
The source-code was provided in a ``source.zip`` file produced with the command :
```bash
zip -r -FS ./source.zip . --exclude 'icons/' --exclude 'node_modules/*' --exclude '.DS_Store' --exclude 'dist/*' --exclude '*.zip' --exclude 'build.local.sh'
```
