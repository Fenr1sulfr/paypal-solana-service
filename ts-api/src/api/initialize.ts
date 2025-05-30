import { Request, Response } from "express";
import { initializePaymentStorage } from "../solana/transactions";
import { PublicKey } from "@solana/web3.js";

export async function handleInitialize(req: Request, res: Response) {
  const { publicKey } = req.body;
  if (!publicKey) return res.status(400).json({ success: false, message: "Missing publicKey" });

  try {
    const owner = new PublicKey(publicKey);
    const tx = await initializePaymentStorage(owner);
    console.log("Transaction created(Init):", tx);
    res.status(200).json({ success: true, transaction: tx });
  } catch (error: any) {
    res.status(500).json({ success: false, message: error.message });
  }
}
