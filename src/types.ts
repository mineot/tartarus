export interface CommandDoc {
  _id: string;
  instructions: string[];
  description?: string;
  _rev?: string;
}

export interface ManualDoc {
  _id: string;
  content: string;
  updatedAt: string;
  _rev?: string;
}

export interface ConfigDoc {
  _id: string;
  data: any;
  updatedAt: string;
  _rev?: string;
}
