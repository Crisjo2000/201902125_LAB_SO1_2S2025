#!/bin/bash

generador_nombre_aleatorio() {
    cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 10 | head -n 1
}

CANT_ARCHIVOS=$((RANDOM % 4 + 1))

echo "Creando $CANT_ARCHIVOS archivos..."

for ((i=1; i<=$CANT_ARCHIVOS; i++)) do
    NOMBRE_ARCHIVO="contenedor_201902125_$(generador_nombre_aleatorio).txt"
    echo "Creando archivo: $NOMBRE_ARCHIVO"
    touch "$NOMBRE_ARCHIVO"
    echo "$NOMBRE_ARCHIVO" > "$NOMBRE_ARCHIVO"
    echo "Archivo $NOMBRE_ARCHIVO creado."
done

