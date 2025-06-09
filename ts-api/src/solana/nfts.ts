import {
  createNft,
  fetchDigitalAsset,
  mplTokenMetadata,
} from "@metaplex-foundation/mpl-token-metadata";
import {
  airdropIfRequired,
  getExplorerLink,
  getKeypairFromFile,
} from "@solana-developers/helpers";
import { createUmi } from "@metaplex-foundation/umi-bundle-defaults";
import { Connection, LAMPORTS_PER_SOL, clusterApiUrl } from "@solana/web3.js";
import {
  generateSigner,
  keypairIdentity,
  percentAmount,
} from "@metaplex-foundation/umi";

async function fetchWithRetry(umi: any, publicKey: any, maxRetries = 10, delay = 2000) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      console.log(`Attempt ${i + 1} to fetch digital asset...`);
      const asset = await fetchDigitalAsset(umi, publicKey);
      return asset;
    } catch (error: any) {
      if (error.message?.includes('AccountNotFoundError') || error.message?.includes('was not found')) {
        console.log(`Asset not found, waiting ${delay}ms before retry ${i + 1}/${maxRetries}...`);
        if (i < maxRetries - 1) {
          await new Promise(resolve => setTimeout(resolve, delay));
          // Increase delay for next retry
          delay = Math.min(delay * 1.2, 10000);
        }
      } else {
        // If it's not an AccountNotFoundError, throw immediately
        throw error;
      }
    }
  }
  throw new Error(`Failed to fetch digital asset after ${maxRetries} attempts`);
}

async function main() {
  try {
    const connection = new Connection(clusterApiUrl("devnet"), "confirmed");
    const user = await getKeypairFromFile();
    
    await airdropIfRequired(
      connection,
      user.publicKey,
      1 * LAMPORTS_PER_SOL,
      0.5 * LAMPORTS_PER_SOL
    );
    
    console.log("Loaded user", user.publicKey.toBase58());
    
    const umi = createUmi(connection.rpcEndpoint);
    umi.use(mplTokenMetadata());
    
    const umiUser = umi.eddsa.createKeypairFromSecretKey(user.secretKey);
    umi.use(keypairIdentity(umiUser));
    
    console.log("Set up Umi instance for user");
    
    const collectionMint = generateSigner(umi);
    
    console.log("Creating NFT collection...");
    const transaction = await createNft(umi, {
      mint: collectionMint,
      name: "Agrichain",
      symbol: "AGCH",
      uri: "https://raw.githubusercontent.com/Fenr1sulfr/paypal-solana-service/main/metadata.json",
      sellerFeeBasisPoints: percentAmount(0),
      isCollection: true,
    });
    
    await transaction.sendAndConfirm(umi);
    console.log("Transaction confirmed, waiting for asset to be available...");
    
    // Use retry logic instead of fixed delay
    const createdCollectionNft = await fetchWithRetry(umi, collectionMint.publicKey);
    
    console.log(
      `Created Collection 📦! Address is ${getExplorerLink(
        "address",
        createdCollectionNft.mint.publicKey,
        "devnet"
      )}`
    );
    
  } catch (error) {
    console.error("Error in main function:", error);
    throw error;
  }
}

main().catch(console.error);