
import { createNftForRecipient } from '../solana/create-nft';
import { Request, Response } from 'express';
import { PublicKey } from '@solana/web3.js';

export const createNftHandler = async (req: Request, res: Response) => {
    try {
        // Extract and validate the recipient public key
        const { recipient } = req.body;
        if (!recipient) {
            return res.status(400).json({ success: false, error: 'Recipient public key is required' });
        }

        let recipientPublicKey: PublicKey;
        try {
            recipientPublicKey = new PublicKey(recipient);
        } catch (e) {
            return res.status(400).json({ success: false, error: 'Invalid recipient public key' });
        }

        // Pass the PublicKey object to the NFT creation function
        const result = await createNftForRecipient(recipient);
        res.status(200).json({ success: true, data: result });
    } catch (error: any) {
        res.status(500).json({ success: false, error: error.message || 'Internal Server Error' });
    }
};