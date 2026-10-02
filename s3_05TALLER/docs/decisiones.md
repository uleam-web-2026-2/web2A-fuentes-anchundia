# Decisiones

## D1 · ¿Quién pone el estado al crear?

**Opciones:** 
Que el cliente envíe el estado al crear el canje, o que el servidor lo asigne automáticamente por seguridad.

**Qué elegimos:** 
Que el servidor imponga el estado inicial de forma automática ("solicitado").

**Por qué, en nuestro negocio:** 
En nuestro modelo de negocio de fidelización por puntos, un cliente no debe tener la capacidad de autoasignarse una recompensa enviando el estado "entregado" en su petición inicial. Por seguridad y control del inventario, el servidor fuerza siempre que un nuevo registro nazca con el estado "solicitado".

**Qué pasaría con la otra opción:** 
Si permitiéramos que el cliente envíe el estado libremente, cualquier usuario podría registrar un canje directamente como "entregado" o "aprobado", saltándose la validación del personal de la cafetería y obteniendo beneficios o productos sin que el sistema verifique sus puntos reales.