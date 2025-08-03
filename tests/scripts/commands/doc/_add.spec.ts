// import { addDocCommand } from 'src/scripts/commands/doc/_add';
// import { CommandDoc } from 'src/types';
// import { Feedback } from 'src/utils/feedback';
// import db from 'test/db';

// jest.mock('test/db');
// jest.mock('src/utils/feedback');

// describe('addDocCommand', () => {
//   const mockCommand: CommandDoc = {
//     _id: 'cmd:test',
//     instructions: ['echo hello'],
//     description: '',
//   };

//   afterAll(async () => {
//     await db.close?.();
//   });

//   beforeEach(() => {
//     jest.clearAllMocks();
//   });

//   it('should add a description to a command', async () => {
//     (db.get as jest.Mock).mockResolvedValue(mockCommand);
//     (db.put as jest.Mock).mockResolvedValue({ ok: true });

//     await addDocCommand('test', 'This is a test command');

//     expect(db.get).toHaveBeenCalledWith('cmd:test');

//     expect(db.put).toHaveBeenCalledWith(
//       expect.objectContaining({
//         description: 'This is a test command',
//       })
//     );

//     expect(Feedback.success).toHaveBeenCalledWith('Description added to "test".');
//   });

//   it('should show notFound message if command does not exist', async () => {
//     (db.get as jest.Mock).mockRejectedValue({ status: 404 });

//     await addDocCommand('ghost', 'desc');

//     expect(Feedback.notFound).toHaveBeenCalledWith('Command "ghost" not found');
//   });

//   it('should show error message for other errors', async () => {
//     (db.get as jest.Mock).mockRejectedValue({ status: 500, message: 'DB failure' });

//     await addDocCommand('fail', 'desc');

//     expect(Feedback.error).toHaveBeenCalledWith('Failed to add description to "fail": DB failure');
//   });
// });
