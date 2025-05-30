import { PublicKey, Connection, LAMPORTS_PER_SOL } from "@solana/web3.js";

export async function checkAccountBalance(connection: Connection, publicKey: PublicKey): Promise<boolean> {
  const balance = await connection.getBalance(publicKey);
  return balance >= 0.001 * LAMPORTS_PER_SOL;
}

export async function findPaymentStoragePda(owner: PublicKey, programId: PublicKey): Promise<[PublicKey, number]> {
  return PublicKey.findProgramAddressSync([Buffer.from("transactionBank"), owner.toBuffer()], programId);
}

export async function findPaymentPda(paymentId: string, owner: PublicKey, programId: PublicKey): Promise<[PublicKey, number]> {
  return PublicKey.findProgramAddressSync([Buffer.from(paymentId), owner.toBuffer()], programId);
}
