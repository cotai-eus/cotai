#!/usr/bin/env bash
set -euo pipefail

info() { printf "[INFO] %s\n" "$*"; }
err()  { printf "[ERROR] %s\n" "$*" >&2; }

REPO_URL="https://github.com/cotai-eus/cotai.git"

usage() {
  cat <<EOF
Usage: $0 [--dry-run|--apply]

Checks the current Git repository and prints recommended commands to add or update
the remote pointing to ${REPO_URL}. By default it does a dry-run (prints commands).

Options:
  --dry-run   (default) Print commands but do not execute them.
  --apply     Execute the chosen command(s) after confirmation.
  -h, --help  Show this help and exit.

Examples:
  $0
  $0 --apply
EOF
}

MODE="dry"
while [[ ${#} -gt 0 ]]; do
  case "$1" in
    --dry-run) MODE="dry"; shift ;;
    --apply) MODE="apply"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) err "Unknown option: $1"; usage; exit 2 ;;
  esac
done

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  err "No parece un repositorio Git en $(pwd)."
  exit 3
fi

CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "")
if [[ -z "$CURRENT_BRANCH" ]]; then
  CURRENT_BRANCH="master"
fi

info "Repositorio Git detectado. Rama actual: $CURRENT_BRANCH"
info "Estado (porcelain):"
git status --porcelain || true

info "Remotos actuales:"
git remote -v || true

EXISTING_ORIGIN_URL=$(git remote get-url origin 2>/dev/null || true)
MATCH_FOUND=false
if git config --local --get-regexp '^remote\.' 2>/dev/null | grep -q "github.com/cotai-eus/cotai"; then
  MATCH_FOUND=true
fi

echo
if [[ "$MATCH_FOUND" == true ]]; then
  info "Ya existe una referencia a github.com/cotai-eus/cotai en la configuración local."
  echo "Comando para verificar exactamente dónde está:"
  echo "  git config --local --get-regexp '^remote\\.' | grep github.com/cotai-eus/cotai"
  exit 0
fi

if [[ -z "$EXISTING_ORIGIN_URL" ]]; then
  info "No existe remote 'origin'. Recomendación: añadir 'origin' apuntando a ${REPO_URL}"
  CMD_ADD="git remote add origin ${REPO_URL}"
  CMD_PUSH="git push -u origin ${CURRENT_BRANCH}"
  echo
  echo "Comandos sugeridos (secuencia):"
  echo "  $CMD_ADD"
  echo "  $CMD_PUSH"
  if [[ "$MODE" == "apply" ]]; then
    read -p "Ejecutar estos comandos? [y/N]: " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
      $CMD_ADD
      $CMD_PUSH
      info "Remote 'origin' añadido y rama $CURRENT_BRANCH empujada con --set-upstream."
    else
      info "Operación cancelada por el usuario."
    fi
  fi
  exit 0
else
  info "Existe remote 'origin' con URL: $EXISTING_ORIGIN_URL"
  if [[ "$EXISTING_ORIGIN_URL" == "$REPO_URL" ]] || [[ "$EXISTING_ORIGIN_URL" == "${REPO_URL%.git}" ]]; then
    info "'origin' ya apunta a ${REPO_URL}. No se requieren cambios.";
    exit 0
  fi

  echo
  info "'origin' apunta a otra URL. Opciones recomendadas:"
  echo
  echo "1) Actualizar 'origin' para que apunte a ${REPO_URL}:"
  echo "   git remote set-url origin ${REPO_URL}"
  echo "   git push -u origin ${CURRENT_BRANCH}"
  echo
  echo "2) Mantener 'origin' y añadir otro remote llamado 'upstream':"
  echo "   git remote add upstream ${REPO_URL}"
  echo "   git push -u upstream ${CURRENT_BRANCH}"
  echo
  echo "3) Renombrar el origin actual y añadir el nuevo origin (preservar histórico):"
  echo "   git remote rename origin origin-old"
  echo "   git remote add origin ${REPO_URL}"
  echo "   git push -u origin ${CURRENT_BRANCH}"

  if [[ "$MODE" == "apply" ]]; then
    echo
    echo "Si eliges ejecutar, responde con el número de la opción a aplicar. Otra entrada cancela."
    read -p "Opción a ejecutar [1/2/3/N]: " opt
    case "$opt" in
      1)
        git remote set-url origin ${REPO_URL}
        git push -u origin ${CURRENT_BRANCH}
        info "Origin actualizado y rama empujada." ;;
      2)
        git remote add upstream ${REPO_URL}
        git push -u upstream ${CURRENT_BRANCH}
        info "Remote 'upstream' añadido y rama empujada." ;;
      3)
        git remote rename origin origin-old
        git remote add origin ${REPO_URL}
        git push -u origin ${CURRENT_BRANCH}
        info "Origin renombrado y nuevo origin añadido." ;;
      *)
        info "Operación cancelada." ;;
    esac
  fi
fi

exit 0
