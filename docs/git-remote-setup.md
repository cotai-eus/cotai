# Script: git-remote-setup.sh

Descripción
-----------
Pequeño helper para comprobar el repositorio Git local y sugerir comandos para añadir o actualizar
un remote apuntando a https://github.com/cotai-eus/cotai.git. Por defecto el script hace un "dry-run" y
muestra los comandos sugeridos.

Uso
---
- Mostrar comandos (dry-run):

```bash
/home/felipe/dev/dev/scripts/git-remote-setup.sh
```

- Ejecutar y aplicar cambios (pide confirmación):

```bash
/home/felipe/dev/dev/scripts/git-remote-setup.sh --apply
```

Qué hace
--------
- Detecta si estás dentro de un repositorio Git.
- Muestra la rama actual y el estado (`git status --porcelain`).
- Muestra remotos actuales (`git remote -v`).
- Si no existe `origin`, sugiere `git remote add origin ...` y `git push -u origin <branch>`.
- Si `origin` apunta a otra URL, sugiere 3 opciones: `set-url`, añadir `upstream`, o renombrar el origin actual.

Notas
-----
- El script no cambia nada por defecto. Usa `--apply` para ejecutar las operaciones; incluso así
  solicita confirmación antes de modificar remotos o empujar ramas.
- Reemplaza `master` por la rama detectada si `git branch --show-current` devuelve el nombre.
