'use strict';

const COMUNIDADES = ['COM-NORTE', 'COM-SUR', 'COM-CENTRO', 'GLOBAL'];
const TIPOS = ['OFERTA', 'DEMANDA'];

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function poissonRandom(lambda) {
  const L = Math.exp(-lambda);
  let p = 1.0;
  let k = 0;

  do {
    k++;
    p *= Math.random();
  } while (p > L);

  return k - 1;
}

module.exports = {
  initPoisson: function (context, events, done) {
    context.vars.inicio = Date.now();
    context.vars.duracion = 30000; // 30 segundos
    context.vars.totalObjetivo = 2000; // 2000 peticiones estocásticas
    context.vars.totalEnviadas = 0;
    context.vars.oleadaRestante = 0;
    context.vars.forzarFinal = false;

    return done();
  },

  configurarOleadaP2P: function (context, events, done) {
    const ahora = Date.now();

    if (!context.vars.forzarFinal && ahora - context.vars.inicio >= context.vars.duracion) {
      const faltantes = context.vars.totalObjetivo - context.vars.totalEnviadas;
      if (faltantes > 0) {
        context.vars.oleadaRestante = faltantes;
        context.vars.forzarFinal = true;
      } else {
        return done(new Error('STOP'));
      }
    }

    if (context.vars.totalEnviadas >= context.vars.totalObjetivo) {
      return done(new Error('STOP'));
    }

    if (!context.vars.forzarFinal && context.vars.oleadaRestante === 0) {
      const restante = context.vars.totalObjetivo - context.vars.totalEnviadas;
      const tamOleada = randomInt(1, Math.min(50, restante));
      context.vars.oleadaRestante = tamOleada;
    }

    const tipo = TIPOS[randomInt(0, TIPOS.length - 1)];
    const comunidad = COMUNIDADES[randomInt(0, COMUNIDADES.length - 1)];
    const prosumidorId = `PROSUMIDOR-${randomInt(1, 500)}`;
    const ordenId = `ORD-${Date.now()}-${randomInt(100, 999)}`;

    // Oferta solar (kWh entre 5 y 100, precio entre 500 y 700 $/kWh)
    const kwh = parseFloat((Math.random() * 95 + 5).toFixed(2));
    const precioKwh = parseFloat((Math.random() * 200 + 500).toFixed(2));

    context.vars.payload = {
      id: ordenId,
      prosumidor_id: prosumidorId,
      comunidad_id: comunidad,
      tipo: tipo,
      kwh: kwh,
      precio_kwh: precioKwh
    };

    context.vars.totalEnviadas++;
    context.vars.oleadaRestante--;

    if (context.vars.forzarFinal && context.vars.oleadaRestante === 0) {
      return done(new Error('STOP'));
    }

    const intervalo = context.vars.forzarFinal
      ? 0
      : poissonRandom(2) * randomInt(1, 10);

    setTimeout(() => done(), intervalo);
  }
};
