import { Request, Response } from "express";
import { addPayment } from "../solana/transactions";
import { PublicKey } from "@solana/web3.js";
import { PaymentData } from "../solana/types";

export async function handleAddPayment(req: Request, res: Response) {
  try {
    const { publicKey, ...paymentData } = req.body;

    if (!publicKey || !paymentData.paymentId) {
      return res.status(400).json({ success: false, message: "Missing required fields" });
    }
    console.log("Received payment data:", paymentData);
    console.log("Received publicKey:", publicKey);
    
    const owner = new PublicKey(publicKey);
    const tx = await addPayment(paymentData as PaymentData, owner);
    console.log("Transaction created(Payment):", tx);
    
    res.status(200).json({ success: true, transaction: tx });
  } catch (error: any) {
    console.error("Error adding payment:", error);
    res.status(500).json({ success: false, message: error.message });
  }
}
