import PouchDB from 'pouchdb';
import path from 'path';
import fs from 'fs';
import os from 'os';

// Define a persistent database path in the user's home directory
const dbPath = path.join(os.homedir(), '.tartarus', 'db', 'commands');

// Ensure the directory exists
fs.mkdirSync(path.dirname(dbPath), { recursive: true });

// Initialize the PouchDB database
const db = new PouchDB(dbPath);

export default db;
