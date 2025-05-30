import { BN } from "@project-serum/anchor";
import { SystemProgram, PublicKey } from "@solana/web3.js";
import { setupProviderAndProgram } from "./program";
import { checkAccountBalance, findPaymentPda, findPaymentStoragePda } from "./utils";
import { PaymentData } from "./types";

export async function initializePaymentStorage(owner: PublicKey): Promise<string> {
  const { provider, program, PROGRAM_ID } = setupProviderAndProgram();
  const [paymentStoragePda] = await findPaymentStoragePda(owner, PROGRAM_ID);
  const connection = provider.connection;

  if (!(await checkAccountBalance(connection, owner))) {
    throw new Error("Insufficient SOL for transaction fees.");
  }

  const tx = await program.methods.initializePaymentsStorage().accounts({
    paymentStorage: paymentStoragePda,
    owner,
    systemProgram: SystemProgram.programId,
  }).transaction();

  const { blockhash } = await connection.getLatestBlockhash("confirmed");
  tx.recentBlockhash = blockhash;
  tx.feePayer = owner;

  return tx.serialize({ requireAllSignatures: false }).toString("base64");
}

export async function addPayment(paymentData: PaymentData, owner: PublicKey): Promise<string> {
  const { provider, program, PROGRAM_ID } = setupProviderAndProgram();
  const [paymentStoragePda] = await findPaymentStoragePda(owner, PROGRAM_ID);
  const [paymentPda] = await findPaymentPda(paymentData.paymentId, owner, PROGRAM_ID);
  const connection = provider.connection;

  const defaultPubkey = new PublicKey("HJyN6Hp7Pz65eqtwJuLoQVaqN5ntTUWYMbK8cdA32P6p");
  const paymentOwner = paymentData.paymentOwner && paymentData.paymentOwner.trim() !== ""
    ? new PublicKey(paymentData.paymentOwner)
    : defaultPubkey;
  const paymentSender = paymentData.paymentSender && paymentData.paymentSender.trim() !== ""
    ? new PublicKey(paymentData.paymentSender)
    : defaultPubkey;

  const tx = await program.methods.addPayment(
    paymentData.paymentId,
    paymentOwner,
    paymentSender,
    new BN(paymentData.paymentTimestamp),
    new BN(paymentData.paymentAmount),
    paymentData.paymentCurrency,
    paymentData.paymentUrl
  ).accounts({
    payment: paymentPda,
    paymentStorage: paymentStoragePda,
    owner,
    systemProgram: SystemProgram.programId,
  }).transaction();

  const { blockhash } = await connection.getLatestBlockhash("confirmed");
  tx.recentBlockhash = blockhash;
  tx.feePayer = owner;

  return tx.serialize({ requireAllSignatures: false }).toString("base64");
}
