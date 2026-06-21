# Mozeidon addon source-code

The source code is written in ``typescript`` and built with ``webpack``.
As specified in the ``package.json``, the ``node`` version should be >= 20 and the ``npm`` version should be >= 10.

Build the artifact (``dist/background.js``) with:

```bash
npm install && npm run build
```

Or, from the repo root: ``just build-firefox``.

Note :
The source-code was provided in a ``source.zip`` file produced with the command :
```bash
zip -r -FS ./source.zip . --exclude 'icons/' --exclude 'node_modules/*' --exclude '.DS_Store' --exclude '*.zip'
```

