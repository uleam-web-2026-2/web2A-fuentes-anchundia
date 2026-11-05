# Frontera HTTP de Canje

Se probó POST /canjes enviando id=777777, estado=confirmado, cliente_id=999999 y puntos_utilizados=1, junto con una recompensa y sucursal válidas. Antes y después devolvió 201 con ID propio, cliente del JWT, costo real y estado solicitado (ver dto_http.json).

El Hito 2 ya usaba EntradaCanje, por eso esos campos ya se ignoraban. El cambio del taller completa la frontera: Canje no tiene etiquetas JSON y todas las respuestas, listas y estructuras anidadas pasan por DTO explícitos. No se atribuye a esta refactorización una vulnerabilidad que el proyecto ya evitaba.

No se añadió PUT: la API del proyecto solo permite el cambio de estado mediante PATCH /canjes/{id}/estado. El servicio fija siempre solicitado al crear, y confirmar/cancelar conserva sus reglas. Las pruebas TestDTO* verifican nombres JSON y ausencia de asociaciones internas en las respuestas.
