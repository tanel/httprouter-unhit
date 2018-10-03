# httprouter-unhit

Drop-in replacement for https://github.com/julienschmidt/httprouter with hit counter, for testing purposes.

Use this package as drop-in replacement for httprouter with hit counter.
Do not use it in production, as it's slow - using mutex and all.
Visit /endpoints or /endpoints/unhit to see hit counter.
