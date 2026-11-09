// TP4: fetch + JSON, nodos del DOM y delegación de eventos.
document.addEventListener("DOMContentLoaded", () => {
    const formZapatilla = document.getElementById("form-zapatilla");
    const formEntrenamiento = document.getElementById("form-entrenamiento");
    const listaZapatillas = document.getElementById("lista-zapatillas");
    const listaEntrenamientos = document.getElementById("lista-entrenamientos");
    const selectZapatilla = document.getElementById("zapatilla");
    const guardarEntrenamiento = document.getElementById("guardar-entrenamiento");
    const inputFecha = document.getElementById("fecha");
    const mensajeGeneral = document.getElementById("mensaje-general");
    const mensajeZapatilla = document.getElementById("mensaje-zapatilla");
    const mensajeEntrenamiento = document.getElementById("mensaje-entrenamiento");
    let zapatillas = [];

    function fechaDeHoy() {
        const hoy = new Date();
        return [hoy.getFullYear(), String(hoy.getMonth() + 1).padStart(2, "0"), String(hoy.getDate()).padStart(2, "0")].join("-");
    }

    function prepararFecha() {
        if (!inputFecha) return;
        inputFecha.max = fechaDeHoy();
        inputFecha.value = inputFecha.max;
    }

    function mostrarMensaje(elemento, texto, esError = false) {
        if (!elemento) return;
        elemento.textContent = texto;
        elemento.className = `message ${esError ? "error" : "success"}`;
        elemento.hidden = texto === "";
    }

    async function pedirJSON(ruta, opciones = {}) {
        const respuesta = await fetch(ruta, opciones);
        if (!respuesta.ok) {
            const detalle = (await respuesta.text()).trim();
            throw new Error(detalle || `La solicitud falló (${respuesta.status}).`);
        }
        // DELETE devuelve 204 y no tiene un cuerpo JSON para decodificar.
        return respuesta.status === 204 ? null : respuesta.json();
    }

    function crearBotonEliminar(id, entidad, descripcion) {
        const boton = document.createElement("button");
        boton.type = "button";
        boton.className = "delete-button";
        boton.textContent = "Eliminar";
        boton.dataset.id = id;
        boton.dataset.entidad = entidad;
        boton.setAttribute("aria-label", `Eliminar ${descripcion}`);
        return boton;
    }

    function mostrarZapatillas() {
        if (!listaZapatillas) return;
        listaZapatillas.textContent = "";

        for (const zapatilla of zapatillas) {
            const item = document.createElement("li");
            item.className = "shoe-item";
            const datos = document.createElement("div");
            const nombre = document.createElement("span");
            nombre.className = "shoe-name";
            nombre.textContent = zapatilla.marca_modelo;
            const kilometros = document.createElement("span");
            kilometros.className = "shoe-distance";
            kilometros.textContent = `${zapatilla.kms_acumulados.toFixed(2)} km acumulados`;
            datos.appendChild(nombre);
            datos.appendChild(kilometros);
            item.appendChild(datos);
            item.appendChild(crearBotonEliminar(zapatilla.id, "zapatillas", zapatilla.marca_modelo));
            listaZapatillas.appendChild(item);
        }

        if (!zapatillas.length) {
            const vacio = document.createElement("li");
            vacio.className = "empty-state";
            vacio.textContent = "Tu próximo recorrido empieza con un par. Agregá tus zapatillas.";
            listaZapatillas.appendChild(vacio);
        }
        document.getElementById("cantidad-zapatillas").textContent = `${zapatillas.length} ${zapatillas.length === 1 ? "par activo" : "pares activos"}`;
        listaZapatillas.setAttribute("aria-busy", "false");
    }

    // La página de entrenamientos necesita las zapatillas para su selector.
    function actualizarSelectorZapatillas() {
        if (!selectZapatilla) return;
        const seleccionAnterior = selectZapatilla.value;
        selectZapatilla.textContent = "";
        const opcionInicial = document.createElement("option");
        opcionInicial.value = "";
        opcionInicial.textContent = zapatillas.length ? "Seleccioná un par" : "Primero agregá una zapatilla";
        selectZapatilla.appendChild(opcionInicial);
        for (const zapatilla of zapatillas) {
            const opcion = document.createElement("option");
            opcion.value = zapatilla.id;
            opcion.textContent = zapatilla.marca_modelo;
            selectZapatilla.appendChild(opcion);
        }
        if (zapatillas.some(zapatilla => String(zapatilla.id) === seleccionAnterior)) {
            selectZapatilla.value = seleccionAnterior;
        }
        selectZapatilla.disabled = zapatillas.length === 0;
        guardarEntrenamiento.disabled = zapatillas.length === 0;
        document.getElementById("ayuda-zapatilla").textContent = zapatillas.length
            ? "Seleccioná el par que usaste en esta salida."
            : "Agregá al menos un par activo para registrar una salida.";
    }

    function agregarCelda(fila, texto, clase = "") {
        const celda = document.createElement("td");
        // Los datos del usuario se insertan como texto, al igual que en la clase.
        celda.textContent = texto;
        celda.className = clase;
        fila.appendChild(celda);
        return celda;
    }

    function mostrarEntrenamientos(entrenamientos) {
        if (!listaEntrenamientos) return;
        listaEntrenamientos.textContent = "";
        for (const entrenamiento of entrenamientos) {
            const fila = document.createElement("tr");
            // Se usa la parte de fecha para evitar cambios de día por zona horaria.
            const fecha = entrenamiento.fecha.slice(0, 10).split("-").reverse().join("/");
            const celdaFecha = agregarCelda(fila, fecha);
            const tipo = document.createElement("span");
            tipo.className = "session-type";
            tipo.textContent = entrenamiento.tipo;
            celdaFecha.appendChild(tipo);
            agregarCelda(fila, `${entrenamiento.distancia_km.toFixed(2)} km`, "number");
            agregarCelda(fila, `${entrenamiento.tiempo_min} min`);
            agregarCelda(fila, `${entrenamiento.ritmo.toFixed(2)} min/km`, "number");
            agregarCelda(fila, `${entrenamiento.calorias} kcal`);
            const zapatilla = zapatillas.find(par => par.id === entrenamiento.zapatilla_id);
            agregarCelda(fila, zapatilla ? zapatilla.marca_modelo : `Par #${entrenamiento.zapatilla_id} (retirado)`);
            agregarCelda(fila, entrenamiento.lugar && entrenamiento.lugar.Valid ? entrenamiento.lugar.String : "—");
            const acciones = agregarCelda(fila, "");
            acciones.appendChild(crearBotonEliminar(entrenamiento.id, "entrenamientos", `entrenamiento del ${fecha}`));
            listaEntrenamientos.appendChild(fila);
        }
        if (!entrenamientos.length) {
            const fila = document.createElement("tr");
            const celda = agregarCelda(fila, "Todavía no hay salidas registradas. Sumá la primera.", "empty-state");
            celda.colSpan = 8;
            listaEntrenamientos.appendChild(fila);
        }
        document.getElementById("cantidad-entrenamientos").textContent = `${entrenamientos.length} ${entrenamientos.length === 1 ? "salida registrada" : "salidas registradas"}`;
        listaEntrenamientos.setAttribute("aria-busy", "false");
    }

    async function actualizarListas() {
        const [pares, salidas] = await Promise.all([
            pedirJSON("/zapatillas"),
            listaEntrenamientos ? pedirJSON("/entrenamientos") : Promise.resolve([])
        ]);
        zapatillas = pares;
        mostrarZapatillas();
        actualizarSelectorZapatillas();
        mostrarEntrenamientos(salidas);
    }

    async function guardar(formulario, ruta, datos, mensaje, despuesDeGuardar) {
        const boton = formulario.querySelector('button[type="submit"]');
        boton.disabled = true;
        mostrarMensaje(mensaje, "");
        mostrarMensaje(mensajeGeneral, "");
        let guardado = false;
        try {
            await pedirJSON(ruta, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(datos)
            });
            guardado = true;
            formulario.reset();
            despuesDeGuardar();
            await actualizarListas();
            mostrarMensaje(mensaje, "Registro guardado correctamente.");
        } catch (error) {
            const detalle = guardado ? `Se guardó el registro, pero no se pudo actualizar la vista. Recargá la página. ${error.message}` : `No se pudo guardar. ${error.message}`;
            mostrarMensaje(mensaje, detalle, true);
        } finally {
            boton.disabled = formulario === formEntrenamiento && zapatillas.length === 0;
        }
    }

    if (formZapatilla) {
        formZapatilla.addEventListener("submit", evento => {
            evento.preventDefault();
            guardar(formZapatilla, "/zapatillas", {
                marca_modelo: document.getElementById("marca-modelo").value.trim(),
                kms_acumulados: Number(document.getElementById("kms-iniciales").value)
            }, mensajeZapatilla, () => {});
        });
    }

    if (formEntrenamiento) {
        formEntrenamiento.addEventListener("submit", evento => {
            evento.preventDefault();
            guardar(formEntrenamiento, "/entrenamientos", {
                fecha: inputFecha.value,
                distancia_km: Number(document.getElementById("distancia").value),
                tiempo_min: Number(document.getElementById("tiempo").value),
                tipo: document.getElementById("tipo").value,
                zapatilla_id: Number(selectZapatilla.value),
                lugar: document.getElementById("lugar").value.trim()
            }, mensajeEntrenamiento, prepararFecha);
        });
    }

    async function eliminarRegistro(evento) {
        // Un listener en cada contenedor atiende los botones creados después del GET.
        const boton = evento.target.closest("button[data-id][data-entidad]");
        if (!boton || !evento.currentTarget.contains(boton)) return;
        boton.disabled = true;
        mostrarMensaje(mensajeGeneral, "");
        let eliminado = false;
        try {
            await pedirJSON(`/${boton.dataset.entidad}/${boton.dataset.id}`, { method: "DELETE" });
            eliminado = true;
            await actualizarListas();
            mostrarMensaje(mensajeGeneral, "Registro eliminado correctamente.");
        } catch (error) {
            const detalle = eliminado ? `Se eliminó el registro, pero no se pudo actualizar la vista. Recargá la página. ${error.message}` : `No se pudo eliminar. ${error.message}`;
            mostrarMensaje(mensajeGeneral, detalle, true);
            boton.disabled = false;
        }
    }

    if (listaZapatillas) listaZapatillas.addEventListener("click", eliminarRegistro);
    if (listaEntrenamientos) listaEntrenamientos.addEventListener("click", eliminarRegistro);
    prepararFecha();
    actualizarListas().catch(error => {
        if (listaZapatillas) {
            listaZapatillas.setAttribute("aria-busy", "false");
            listaZapatillas.querySelector(".empty-state").textContent = "No se pudo cargar el calzado.";
        }
        if (listaEntrenamientos) {
            listaEntrenamientos.setAttribute("aria-busy", "false");
            listaEntrenamientos.querySelector(".empty-state").textContent = "No se pudo cargar el historial.";
        }
        mostrarMensaje(mensajeGeneral, `No se pudieron cargar los registros. ${error.message}`, true);
    });
});
