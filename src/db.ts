import PouchDB from 'pouchdb';
import path from 'path';
import fs from 'fs';
import os from 'os';

const dbPath = path.join(os.homedir(), '.tartarus', 'db', 'commands');
fs.mkdirSync(path.dirname(dbPath), { recursive: true });

const db = new PouchDB(dbPath);
export default db;
