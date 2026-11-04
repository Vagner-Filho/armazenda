/**
 * Contact fixtures for the detached NF-e e2e specs.
 *
 * Creates a farm contact (person + legal/natural person + optional address)
 * directly in the e2e database and removes it in dependency order. Used by the
 * transportadora selector and recipient farm-scope cases.
 */

const { execSync } = require('child_process');

const DB = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

/** Runs a query against the e2e database and returns its first output line. */
function db(sql) {
  return execSync(
    `PGPASSWORD=${DB.password} psql -v ON_ERROR_STOP=1 -h ${DB.host} -p ${DB.port} -U ${DB.user} -d ${DB.database} -tAc "${sql.replace(/"/g, '\\"')}"`,
    { encoding: 'utf8' }
  ).trim().split('\n')[0];
}

/** Escapes a value for a single-quoted SQL literal. */
function quote(value) {
  return `'${String(value).replace(/'/g, "''")}'`;
}

/**
 * Creates a contact.
 * @param {object} options
 * @param {number} options.farmId farm id the contact belongs to.
 * @param {'legal'|'natural'} [options.kind] person kind.
 * @param {string} options.name company or person name.
 * @param {string} options.document CPF (natural) or CNPJ (legal).
 * @param {string} [options.ie] inscrição estadual (defaults to blank).
 * @param {object} [options.address] optional address fields:
 *   street, cep, number, neighborhood, city, state.
 * @returns {number} the person id.
 */
function createContact({ farmId, kind = 'legal', name, document, ie = '', address }) {
  const personID = Number(
    db(`INSERT INTO person (ie, farm) VALUES (${quote(ie)}, ${farmId}) RETURNING id`)
  );
  if (!personID) throw new Error('could not create the contact person row');

  if (kind === 'natural') {
    db(
      `INSERT INTO natural_person (name, cpf, personId) VALUES (${quote(name)}, ${quote(document)}, ${personID})`
    );
  } else {
    db(
      `INSERT INTO legal_person (cnpj, personId, companyName) VALUES (${quote(document)}, ${personID}, ${quote(name)})`
    );
  }

  if (address) {
    const number = address.number === undefined || address.number === null ? 'NULL' : address.number;
    db(
      `INSERT INTO address (street, cep, number, neighborhood, city, state, person_id)
       VALUES (${quote(address.street)}, ${quote(address.cep)}, ${number}, ${quote(address.neighborhood)}, ${quote(address.city)}, ${quote(address.state)}, ${personID})`
    );
  }

  return personID;
}

/** Deletes a contact and its dependent rows. */
function deleteContact(personID) {
  db(`DELETE FROM address WHERE person_id = ${personID}`);
  db(`DELETE FROM legal_person WHERE personId = ${personID}`);
  db(`DELETE FROM natural_person WHERE personId = ${personID}`);
  db(`DELETE FROM person WHERE id = ${personID}`);
}

module.exports = {
  db,
  createContact,
  deleteContact,
};
