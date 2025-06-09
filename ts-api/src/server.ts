import express from "express";
import cors from "cors";
import { handleInitialize } from "./api/initialize";
import { handleAddPayment } from "./api/addPayment";
import SolanaPayHandler, { SolanaVerifyHandler } from "./api/payment";
import { createNftHandler } from "./api/nft";

const app = express();
app.use(cors());
app.use(express.json());

app.post("/initialize", handleInitialize);
app.post("/add-payment", handleAddPayment);
app.post("/api/pay",SolanaPayHandler);
app.get("/api/pay",SolanaVerifyHandler);
app.post("/nft/create",createNftHandler);
const PORT = process.env.PORT || 8081;
app.listen(PORT, () => console.log(`Server is running on http://localhost:${PORT}`));
