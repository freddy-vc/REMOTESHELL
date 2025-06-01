package main

import "sync"

// Mutex compartido para sincronizar el acceso a la conexión
var ResponseMutex = &sync.Mutex{}
