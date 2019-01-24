var Moniker = require('./moniker');

var names = Moniker.generator([Moniker.verb, Moniker.adjective, Moniker.noun]);

console.log(names.choose(process.env.SEED || ""));