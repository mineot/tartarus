import { CommandDoc } from 'src/types';
import { Feedback } from 'src/utils/feedback';
import { removeDocCommand } from 'src/scripts/commands/doc/_remove';
import db from 'test/db';

jest.mock('test/db');
jest.mock('src/utils/feedback');

describe('removeDocCommand', () => {
  const baseDoc: CommandDoc = {
    _id: 'cmd:test',
    instructions: ['echo test'],
    description: 'A command to test',
  };

  afterAll(async () => {
    await db.close?.();
  });

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should remove description and update the command', async () => {
    (db.get as jest.Mock).mockResolvedValue({ ...baseDoc });
    (db.put as jest.Mock).mockResolvedValue({ ok: true });

    await removeDocCommand('test');

    expect(db.get).toHaveBeenCalledWith('cmd:test');
    expect(db.put).toHaveBeenCalledWith(
      expect.not.objectContaining({ description: expect.any(String) })
    );
    expect(Feedback.success).toHaveBeenCalledWith('Description removed from "test".');
  });

  it('should warn if there is no description to remove', async () => {
    const noDescDoc = { ...baseDoc };
    delete noDescDoc.description;

    (db.get as jest.Mock).mockResolvedValue(noDescDoc);

    await removeDocCommand('test');

    expect(Feedback.warn).toHaveBeenCalledWith('No description to remove from "test".');
    expect(db.put).not.toHaveBeenCalled();
  });

  it('should handle command not found', async () => {
    (db.get as jest.Mock).mockRejectedValue({ status: 404 });

    await removeDocCommand('ghost');

    expect(Feedback.notFound).toHaveBeenCalledWith('Command "ghost" not found');
  });

  it('should handle other errors', async () => {
    (db.get as jest.Mock).mockRejectedValue({ status: 500, message: 'DB failure' });

    await removeDocCommand('fail');

    expect(Feedback.error).toHaveBeenCalledWith(
      'Failed to remove description from "fail": DB failure'
    );
  });
});
