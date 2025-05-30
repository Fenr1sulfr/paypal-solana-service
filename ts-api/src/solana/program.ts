// import { AnchorProvider, Wallet, web3, Program } from "@project-serum/anchor";
import * as anchor from "@project-serum/anchor";
import { PublicKey } from "@solana/web3.js";
import idl from "../idl/receipt.json";

const PROGRAM_ID = new PublicKey("YRoyDgtmHvKDpFdksFPdcCB16ymBspq2kUhgVz18JFQ");
const IDL = idl as anchor.Idl; // Cast to Idl type

export function setupProviderAndProgram() {
  const connection = new anchor.web3.Connection(anchor.web3.clusterApiUrl("devnet"), "confirmed");
  const dummyKeypair = anchor.web3.Keypair.generate();
  const provider = new anchor.AnchorProvider(connection, new anchor.Wallet(dummyKeypair), { preflightCommitment: "confirmed" });
  const program = new anchor.Program(IDL, PROGRAM_ID, provider);
  return { provider, program, connection, PROGRAM_ID };
}
