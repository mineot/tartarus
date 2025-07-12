import PouchDB from 'pouchdb';
import memory from 'pouchdb-adapter-memory';

PouchDB.plugin(memory);

const db = new PouchDB('tartarus-test', { adapter: 'memory' });

export default db;
