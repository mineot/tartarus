import { CommandDoc } from './types';
import path from 'path';
import PouchDB from 'pouchdb';

const dbPath = path.join(__dirname, '..', 'db', 'commands');
const db = new PouchDB<CommandDoc>(dbPath);

export default db;
