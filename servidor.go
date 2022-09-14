package main

import (
	"fmt"      // Imprimir en consola
	"io"       // Ayuda a escribir en la respuesta
	"log"      //Loguear si algo sale mal
	"net/http" // El paquete HTTP
	"os"
	"path/filepath"

	"github.com/rs/xid"
)

func crearDirectorioSiNoExiste(directorio string) error {
	if _, err := os.Stat(directorio); os.IsNotExist(err) {
		err = os.Mkdir(directorio, 0755)
		if err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func obtenerIdAleatorioNoSeguro() string {
	guid := xid.New()
	return guid.String()
}

func renombrarNombreDeArchivoAIdAleatorio(nombreOriginal string) string {
	extension := filepath.Ext(nombreOriginal)
	return obtenerIdAleatorioNoSeguro() + extension
}

func main() {

	http.HandleFunc("/foto", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			io.WriteString(w, "Solo se permiten peticiones POST")
			return
		}

		const MaximoTamanioFotosEnBytes = 5 << 20 // 5 megabytes, recuerda que debe haber espacio para tamaño foto + datos adicionales (o sea, formulario)
		const DirectorioArchivos = "subidas"
		crearDirectorioSiNoExiste(DirectorioArchivos)
		// CORS
		const HostPermitidoParaCORS = "http://localhost"

		w.Header().Set("Access-Control-Allow-Origin", HostPermitidoParaCORS)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Prevenir que envíen peticiones muy grandes. Recuerda dejar espacio para máximo tamaño de foto + datos adicionales
		r.Body = http.MaxBytesReader(w, r.Body, MaximoTamanioFotosEnBytes)
		// Parsea y aloja en RAM o disco duro, dependiendo del límite que le indiquemos
		err := r.ParseMultipartForm(MaximoTamanioFotosEnBytes)
		if err != nil {
			log.Printf("Error al parsear: %v", err)
			return
		}
		encabezadosDeArchivos := r.MultipartForm.File["archivo"]

		nombre := r.Form.Get("nombre")
		log.Printf("Nombre: %v", nombre)

		encabezadoPrimerArchivo := encabezadosDeArchivos[0]
		primerArchivo, err := encabezadoPrimerArchivo.Open()
		if err != nil {
			log.Printf("Error al abrir archivo: %v", err)
			return
		}
		defer primerArchivo.Close()
		nombreArchivo := renombrarNombreDeArchivoAIdAleatorio(encabezadoPrimerArchivo.Filename)
		archivoParaGuardar, err := os.Create(filepath.Join(DirectorioArchivos, nombreArchivo))
		if err != nil {
			return
		}
		defer archivoParaGuardar.Close()
		_, err = io.Copy(archivoParaGuardar, primerArchivo)
		if err != nil {
			return
		}
		io.WriteString(w, "Subido correctamente")
	})

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	direccion := ":8080" // Como cadena, no como entero; porque representa una dirección
	fmt.Println("Servidor listo escuchando en " + direccion)
	log.Fatal(http.ListenAndServe(direccion, nil))
}
